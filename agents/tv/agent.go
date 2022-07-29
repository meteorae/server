package tv

import (
	"fmt"
	"time"

	"github.com/agnivade/levenshtein"
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
	return "TV Agent"
}

// TODO: Currently this returns only the first result, since we can't fix mismatches anyway.
func getTVSeriesResults(item database.ItemMetadata) (tmdb.TV, error) {
	options := map[string]string{
		"language":      "en-US", // Make this configurable
		"include_adult": "false", // Make this configurable
	}

	if item.ReleaseDate.Year() > 0 {
		options["year"] = fmt.Sprintf("%d", item.ReleaseDate.Year())
	}

	searchResults, err := tmdbAPI.SearchTv(item.Title, options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for series %s", item.Title)

		return tmdb.TV{}, err
	}

	// TODO: Add other sorting methods to complement the Levenshtein distance.
	// We should score results out of 100 and return all of them (sorted by score).
	// Possible candidates are (based on data we have during scans):
	// - Sorting results by popularity (Most popular titles have more chance to be what the user is looking for)
	// - Compare release date (Or the year, if we only have that)
	if len(searchResults.Results) > 0 {
		// Calculate the levenshtein distance between the title and the titles of all the results
		// and return the result with the lowest distance.
		var (
			minDistance      int
			minDistanceIndex int
		)

		for i, result := range searchResults.Results {
			distance := levenshtein.ComputeDistance(item.Title, result.Name)
			if i == 0 || distance < minDistance {
				minDistance = distance
				minDistanceIndex = i
			}
		}

		resultMovie := searchResults.Results[minDistanceIndex]

		seriesData, err := tmdbAPI.GetTvInfo(resultMovie.ID, map[string]string{})
		if err != nil {
			log.Err(err).Msgf("failed to fetch information for series \"%s\": %w", item.Title, err)
		}

		return *seriesData, nil
	}

	return tmdb.TV{}, errNoResultsFound
}

func Search(item database.ItemMetadata) {
	// Series
	seriesInfo, err := getTVSeriesResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.Title)

		return
	}

	releaseDate, err := time.Parse("2006-01-02", seriesInfo.FirstAirDate)
	if err != nil {
		log.Err(err).Msgf("Failed to parse release date for series \"%s\"", item.Title)

		releaseDate = time.Time{}
	}

	var artHash string

	if seriesInfo.BackdropPath != "" {
		artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", seriesInfo.BackdropPath)

		artHash, err = helpers.SaveExternalImageToCache(artPath)
		if err != nil {
			log.Err(err).Msgf("Failed to download backdrop for series \"%s\"", item.Title)
		}
	}

	var posterHash string

	if seriesInfo.PosterPath != "" {
		posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", seriesInfo.PosterPath)

		posterHash, err = helpers.SaveExternalImageToCache(posterPath)
		if err != nil {
			log.Err(err).Msgf("failed to download poster for series \"%s\"", item.Title)
		}
	}

	err = item.Update(database.ItemMetadata{
		Title:       seriesInfo.Name,
		SortTitle:   utils.CleanSortTitle(seriesInfo.Name),
		ReleaseDate: releaseDate,
		Summary:     seriesInfo.Overview,
		Thumb:       posterHash,
		Art:         artHash,
	})
	if err != nil {
		log.Err(err).Msgf("Failed to update series %s", item.Title)
	}
}
