package movie

import (
	"errors"
	"fmt"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
	"github.com/ryanbradynd05/go-tmdb"
	"golang.org/x/text/language"
)

var (
	errNoResultsFound = fmt.Errorf("no results found")
	apiKey            = "c9ae218044f9b20a4fcbba36d543a730" //#nosec
	config            = tmdb.Config{
		APIKey:   apiKey,
		Proxies:  nil,
		UseProxy: false,
	}
)

var tmdbAPI *tmdb.TMDb = tmdb.Init(config)

func GetName() string {
	return "Meteorae Movie Agent"
}

func GetSearchResults(item database.ItemMetadata) ([]sdk.Item, error) {
	options := map[string]string{
		"language":      "en-US", // TODO: Make this configurable
		"include_adult": "false", // TODO: Make this configurable
	}

	if !item.ReleaseDate.IsZero() {
		options["year"] = fmt.Sprintf("%d", item.ReleaseDate.Year())
	}

	searchResults, err := tmdbAPI.SearchMovie(item.Title, options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.Title)

		return []sdk.Item{}, err
	}

	results := make([]sdk.Item, 0, len(searchResults.Results))

	for _, result := range searchResults.Results {
		releaseDate, err := time.Parse("2006-01-02", result.ReleaseDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", result.Title)

			releaseDate = time.Time{}
		}

		results = append(results, sdk.Movie{
			ItemInfo: &sdk.ItemInfo{
				Title:         result.Title,
				OriginalTitle: result.OriginalTitle,
				ReleaseDate:   releaseDate,
			},
			Popularity: result.Popularity,
			TmdbID:     result.ID,
		})
	}

	if len(results) > 0 {
		return results, nil
	}

	return []sdk.Item{}, errNoResultsFound
}

func GetMetadata(item database.ItemMetadata) (database.ItemMetadata, error) {
	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to get movie results %s", item.Title)

		return database.ItemMetadata{}, err
	}

	if len(results) == 0 {
		return database.ItemMetadata{}, nil
	}

	resultMovie := results[0]

	if media, ok := resultMovie.(sdk.Movie); ok {
		movieData, err := tmdbAPI.GetMovieInfo(media.TmdbID, map[string]string{})
		if err != nil {
			log.Err(err).Msgf("failed to fetch information for movie \"%s\"", item.Title)
		}

		releaseDate, err := time.Parse("2006-01-02", movieData.ReleaseDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", movieData.Title)

			releaseDate = time.Time{}
		}

		languageTag, err := language.Parse(movieData.OriginalLanguage)
		if err != nil {
			log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", movieData.Title)

			languageTag = language.Und
		}

		var artHash string

		if movieData.BackdropPath != "" {
			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

			artHash, err = helpers.SaveExternalImageToCache(artPath)
			if err != nil {
				log.Err(err).Msgf("Failed to download backdrop for movie \"%s\"", item.Title)
			}
		}

		var posterHash string

		if movieData.PosterPath != "" {
			posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

			posterHash, err = helpers.SaveExternalImageToCache(posterPath)
			if err != nil {
				log.Err(err).Msgf("failed to download poster for movie \"%s\"", item.Title)
			}
		}

		return database.ItemMetadata{
			Title:            movieData.Title,
			OriginalTitle:    movieData.OriginalTitle,
			SortTitle:        utils.CleanSortTitle(movieData.Title),
			ReleaseDate:      releaseDate,
			Tagline:          movieData.Tagline,
			Summary:          movieData.Overview,
			OriginalLanguage: languageTag.String(),
			Thumb:            posterHash,
			Art:              artHash,
		}, nil
	}

	return database.ItemMetadata{}, errors.New("unsupported item type")
}
