package movie

import (
	"fmt"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
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
	return "Movie Agent"
}

// TODO: Currently this returns only the first result, since we can't fix mismatches anyway.
func getMovieResults(media models.Movie) (tmdb.Movie, error) {
	options := map[string]string{
		"language":      "en-US", // Make this configurable
		"include_adult": "false", // Make this configurable
	}

	if media.ReleaseDate.Year() > 0 {
		options["year"] = fmt.Sprintf("%d", media.ReleaseDate.Year())
	}

	searchResults, err := tmdbAPI.SearchMovie(media.Title, options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", media.Title)

		return tmdb.Movie{}, err
	}

	if len(searchResults.Results) > 0 {
		resultMovie := searchResults.Results[0]

		movieData, err := tmdbAPI.GetMovieInfo(resultMovie.ID, map[string]string{})
		if err != nil {
			log.Err(err).Msgf("failed to fetch information for movie \"%s\": %w", media.Title, err)
		}

		return *movieData, nil
	}

	return tmdb.Movie{}, errNoResultsFound
}

func Search(media database.ItemMetadata) {
	movie, err := getMovieResults(media)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", media.Title)

		return
	}

	releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
	if err != nil {
		log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", media.Title)

		releaseDate = time.Time{}
	}

	languageTag, err := language.Parse(movie.OriginalLanguage)
	if err != nil {
		log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", media.Title)

		languageTag = language.Und
	}

	var artHash string

	if movie.BackdropPath != "" {
		artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movie.BackdropPath)

		artHash, err = helpers.SaveExternalImageToCache(artPath)
		if err != nil {
			log.Err(err).Msgf("Failed to download backdrop for movie \"%s\"", media.Title)
		}
	}

	var posterHash string

	if movie.PosterPath != "" {
		posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movie.PosterPath)

		posterHash, err = helpers.SaveExternalImageToCache(posterPath)
		if err != nil {
			log.Err(err).Msgf("failed to download poster for movie \"%s\"", media.Title)
		}
	}

	media.Title = movie.Title
	media.SortTitle = utils.CleanSortTitle(movie.Title)
	media.OriginalTitle = movie.OriginalTitle
	media.ReleaseDate = releaseDate
	media.Summary = movie.Overview
	media.Tagline = movie.Tagline
	media.Popularity = movie.Popularity
	media.OriginalLanguage = languageTag.String()
	media.Thumb = posterHash
	media.Art = artHash

	err = database.UpdateMovie(&media)
	if err != nil {
		log.Err(err).Msgf("Failed to update movie %s", media.Title)
	}
}
