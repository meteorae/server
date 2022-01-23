package themoviedb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/adrg/xdg"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/utils"
	PTN "github.com/middelink/go-parse-torrent-name"
	"github.com/rs/zerolog/log"
	"github.com/ryanbradynd05/go-tmdb"
	"golang.org/x/text/language"
	"gopkg.in/gographics/imagick.v2/imagick"
)

var errNoResultsFound = fmt.Errorf("no results found")

var apiKey = "c9ae218044f9b20a4fcbba36d543a730"

var config = tmdb.Config{
	APIKey:   apiKey,
	Proxies:  nil,
	UseProxy: false,
}

var tmdbAPI *tmdb.TMDb = tmdb.Init(config)

func GetMovieInfoFromTmdb(movie *PTN.TorrentInfo, mediaPart *models.MediaPart) (*models.ItemMetadata, error) {
	// Remove unwanted characters from the title
	movie.Title = utils.RemoveUnwantedCharacters(movie.Title)

	// Some movies have multiple languages or versions in the name using "aka", get only the first one
	reg := regexp.MustCompile("(.*) aka .*")
	cleanTitle := reg.FindStringSubmatch(movie.Title)

	if len(cleanTitle) > 0 {
		log.Info().Msgf("Title cleaned up to %s", cleanTitle[1])
		movie.Title = cleanTitle[1]
	}

	searchResults, err := tmdbAPI.SearchMovie(movie.Title, map[string]string{
		"year":          fmt.Sprintf("%d", movie.Year),
		"language":      "en-US",
		"include_adult": "false",
	})
	if err != nil {
		return nil, fmt.Errorf("could not search for movie: %w", err)
	}

	if len(searchResults.Results) > 0 {
		resultMovie := searchResults.Results[0]

		movieData, err := tmdbAPI.GetMovieInfo(resultMovie.ID, map[string]string{})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch information for movie \"%s\": %w", movie.Title, err)
		}

		releaseDate, err := time.Parse("2006-01-02", movieData.ReleaseDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", movie.Title)

			releaseDate = time.Time{}
		}

		languageTag, err := language.Parse(movieData.OriginalLanguage)
		if err != nil {
			log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", movie.Title)

			languageTag = language.Und
		}

		magickWand := imagick.NewMagickWand()
		defer magickWand.Destroy()

		var artHash string
		if movieData.BackdropPath != "" {
			var artBuffer bytes.Buffer

			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

			response, err := http.Get(artPath)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch art for movie \"%s\": %w", movie.Title, err)
			}
			defer response.Body.Close()

			_, err = io.Copy(&artBuffer, response.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to copy art for movie \"%s\": %w", movie.Title, err)
			}

			hash, err := utils.HashFileBytes(artBuffer.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to hash art for movie \"%s\": %w", movie.Title, err)
			}

			artHash = hex.EncodeToString(hash)
			prefix := artHash[0:2]

			imageCachePath, err := xdg.CacheFile("meteorae/images")
			if err != nil {
				return nil, fmt.Errorf("failed to get image cache path: %w", err)
			}

			err = magickWand.ReadImageBlob(artBuffer.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to read art for movie \"%s\": %w", movie.Title, err)
			}

			err = magickWand.SetImageFormat("webp")
			if err != nil {
				return nil, fmt.Errorf("failed to set image format: %w", err)
			}

			filePath := filepath.Join(imageCachePath, prefix, artHash)

			err = os.MkdirAll(filePath, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create image cache directory: %w", err)
			}

			filePath = filepath.Join(filePath, "0x0.webp")

			err = magickWand.WriteImage(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to write image to disk: %w", err)
			}
		}

		var posterHash string
		if movieData.PosterPath != "" {
			var posterBuffer bytes.Buffer

			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

			response, err := http.Get(artPath)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch poster for movie \"%s\": %w", movie.Title, err)
			}
			defer response.Body.Close()

			_, err = io.Copy(&posterBuffer, response.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to copy poster for movie \"%s\": %w", movie.Title, err)
			}

			hash, err := utils.HashFileBytes(posterBuffer.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to hash poster for movie \"%s\": %w", movie.Title, err)
			}

			posterHash = hex.EncodeToString(hash)
			prefix := posterHash[0:2]

			imageCachePath, err := xdg.CacheFile("meteorae/images")
			if err != nil {
				return nil, fmt.Errorf("failed to get image cache path: %w", err)
			}

			err = magickWand.ReadImageBlob(posterBuffer.Bytes())
			if err != nil {
				return nil, fmt.Errorf("failed to read poster for movie \"%s\": %w", movie.Title, err)
			}

			err = magickWand.SetImageFormat("webp")
			if err != nil {
				return nil, fmt.Errorf("failed to set image format: %w", err)
			}

			filePath := filepath.Join(imageCachePath, prefix, posterHash)

			err = os.MkdirAll(filePath, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create image cache directory: %w", err)
			}

			filePath = filepath.Join(filePath, "0x0.webp")

			err = magickWand.WriteImage(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to write image to disk: %w", err)
			}
		}

		return &models.ItemMetadata{
			Title:            movieData.Title,
			SortTitle:        utils.CleanSortTitle(movieData.Title),
			OriginalTitle:    movieData.OriginalTitle,
			ReleaseDate:      releaseDate,
			Tagline:          movieData.Tagline,
			Popularity:       movieData.Popularity,
			OriginalLanguage: languageTag.String(),
			Thumb:            posterHash,
			Art:              artHash,
			MediaPart:        *mediaPart,
		}, nil
	}

	return nil, errNoResultsFound
}
