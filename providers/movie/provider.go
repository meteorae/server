package movie

import (
	"fmt"
	"regexp"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
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

func GetInformation(item *database.ItemMetadata, library database.Library) error {
	// Remove unwanted characters from the title
	item.Title = utils.RemoveUnwantedCharacters(item.Title)

	// Some movies have multiple languages or versions in the name using "aka", get only the first one
	reg := regexp.MustCompile("(.*) aka .*")
	cleanTitle := reg.FindStringSubmatch(item.Title)

	if len(cleanTitle) > 0 {
		log.Info().Msgf("Title cleaned up to %s", cleanTitle[1])
		item.Title = cleanTitle[1]
	}

	searchResults, err := tmdbAPI.SearchMovie(item.Title, map[string]string{
		"language":      "en-US", // Make this configurable
		"include_adult": "false", // Make this configurable
	})
	if err != nil {
		return fmt.Errorf("could not search for movie: %w", err)
	}

	if len(searchResults.Results) > 0 {
		resultMovie := searchResults.Results[0]

		movieData, err := tmdbAPI.GetMovieInfo(resultMovie.ID, map[string]string{})
		if err != nil {
			return fmt.Errorf("failed to fetch information for movie \"%s\": %w", item.Title, err)
		}

		releaseDate, err := time.Parse("2006-01-02", movieData.ReleaseDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", item.Title)

			releaseDate = time.Time{}
		}

		languageTag, err := language.Parse(movieData.OriginalLanguage)
		if err != nil {
			log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", item.Title)

			languageTag = language.Und
		}

		var artHash string

		if movieData.BackdropPath != "" {
			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

			artHash, err = helpers.SaveExternalImageToCache(artPath)
			if err != nil {
				return fmt.Errorf("failed to download backdrop art for movie \"%s\": %w", item.Title, err)
			}
		}

		var posterHash string

		if movieData.PosterPath != "" {
			posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

			posterHash, err = helpers.SaveExternalImageToCache(posterPath)
			if err != nil {
				return fmt.Errorf("failed to download poster art for movie \"%s\": %w", item.Title, err)
			}
		}

		updates := map[string]interface{}{
			"title":            movieData.Title,
			"sortTitle":        utils.CleanSortTitle(movieData.Title),
			"originalTitle":    movieData.OriginalTitle,
			"releaseDate":      releaseDate,
			"summary":          movieData.Overview,
			"tagline":          movieData.Tagline,
			"popularity":       movieData.Popularity,
			"originalLanguage": languageTag.String(),
			"thumb":            posterHash,
			"art":              artHash,
		}

		database.UpdateItem(item.Id, updates)
	}

	return errNoResultsFound
}
