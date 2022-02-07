package themoviedb

import (
	"fmt"
	"regexp"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
	PTN "github.com/middelink/go-parse-torrent-name"
	"github.com/rs/zerolog/log"
	"github.com/ryanbradynd05/go-tmdb"
	"golang.org/x/text/language"
)

var errNoResultsFound = fmt.Errorf("no results found")

var apiKey = "c9ae218044f9b20a4fcbba36d543a730" //#nosec

var config = tmdb.Config{
	APIKey:   apiKey,
	Proxies:  nil,
	UseProxy: false,
}

var tmdbAPI *tmdb.TMDb = tmdb.Init(config)

func GetMovieInfoFromTmdb(movie *PTN.TorrentInfo, mediaPart *database.MediaPart,
	library database.Library) (*database.ItemMetadata, error) {
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

		var artHash string

		if movieData.BackdropPath != "" {
			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

			artHash, err = helpers.SaveExternalImageToCache(artPath)
			if err != nil {
				return nil, fmt.Errorf("failed to download backdrop art for movie \"%s\": %w", movie.Title, err)
			}
		}

		var posterHash string

		if movieData.PosterPath != "" {
			posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

			posterHash, err = helpers.SaveExternalImageToCache(posterPath)
			if err != nil {
				return nil, fmt.Errorf("failed to download poster art for movie \"%s\": %w", movie.Title, err)
			}
		}

		return &database.ItemMetadata{
			Title:            movieData.Title,
			SortTitle:        utils.CleanSortTitle(movieData.Title),
			OriginalTitle:    movieData.OriginalTitle,
			ReleaseDate:      releaseDate,
			Summary:          movieData.Overview,
			Tagline:          movieData.Tagline,
			Popularity:       movieData.Popularity,
			OriginalLanguage: languageTag.String(),
			Thumb:            posterHash,
			Art:              artHash,
			MediaPart:        *mediaPart,
			Library:          library,
		}, nil
	}

	return nil, errNoResultsFound
}
