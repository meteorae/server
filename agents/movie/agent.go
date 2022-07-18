package movie

import (
	"fmt"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
	"github.com/ryanbradynd05/go-tmdb"
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
	return "Movie Agent"
}

// TODO: Currently this returns only the first result, since we can't fix mismatches anyway.
func getMovieResults(item database.ItemMetadata) (tmdb.Movie, error) {
	options := map[string]string{
		"language":      "en-US", // Make this configurable
		"include_adult": "false", // Make this configurable
	}

	if item.ReleaseDate.Year() > 0 {
		options["year"] = fmt.Sprintf("%d", item.ReleaseDate.Year())
	}

	searchResults, err := tmdbAPI.SearchMovie(item.Title, options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.Title)

		return tmdb.Movie{}, err
	}

	if len(searchResults.Results) > 0 {
		resultMovie := searchResults.Results[0]

		movieData, err := tmdbAPI.GetMovieInfo(resultMovie.ID, map[string]string{})
		if err != nil {
			log.Err(err).Msgf("failed to fetch information for movie \"%s\": %w", item.Title, err)
		}

		return *movieData, nil
	}

	return tmdb.Movie{}, errNoResultsFound
}

func Search(item database.ItemMetadata) {
	movie, err := getMovieResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.Title)

		return
	}

	releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
	if err != nil {
		log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", item.Title)

		releaseDate = time.Time{}
	}

	/*languageTag, err := language.Parse(movie.OriginalLanguage)
	if err != nil {
		log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", media.Title)

		languageTag = language.Und
	}*/

	var artHash string

	if movie.BackdropPath != "" {
		artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movie.BackdropPath)

		artHash, err = helpers.SaveExternalImageToCache(artPath)
		if err != nil {
			log.Err(err).Msgf("Failed to download backdrop for movie \"%s\"", item.Title)
		}
	}

	var posterHash string

	if movie.PosterPath != "" {
		posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movie.PosterPath)

		posterHash, err = helpers.SaveExternalImageToCache(posterPath)
		if err != nil {
			log.Err(err).Msgf("failed to download poster for movie \"%s\"", item.Title)
		}
	}

	item.Title = movie.Title
	item.SortTitle = utils.CleanSortTitle(movie.Title)
	item.ReleaseDate = releaseDate
	item.Summary = movie.Overview
	item.Thumb = posterHash
	item.Art = artHash

	err = item.Update()
	if err != nil {
		log.Err(err).Msgf("Failed to update movie %s", item.Title)
	}

	for _, observer := range utils.SubsciptionsManager.ItemUpdatedObservers {
		observer <- helpers.GetItemFromItemMetadata(item)
	}
}
