package tvshow

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
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
	yearRegex   = regexp.MustCompile(`([ ]+\(?[0-9]{4}\)?)`)
	prefixRegex = regexp.MustCompile(`^[Bb][Bb][Cc] `)
)

var tmdbAPI *tmdb.TMDb = tmdb.Init(config)

func GetName() string {
	return "Meteorae TV Agent"
}

// FIXME: Kingdom (2019) is matching to Kingdom (2017). Seems like there's an issue with the
//        date or we're incorrectly lowering the proper match's score.
func GetSearchResults(item database.ItemMetadata) ([]sdk.Item, error) {
	options := map[string]string{
		"language":      "en-US", // TODO: Make this configurable
		"include_adult": "false", // TODO: Make this configurable
	}

	if !item.ReleaseDate.IsZero() {
		options["year"] = fmt.Sprintf("%d", item.ReleaseDate.Year())
	}

	searchResults, err := tmdbAPI.SearchTv(item.Title, options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for series %s", item.Title)

		return []sdk.Item{}, err
	}

	results := make([]sdk.Item, 0, len(searchResults.Results))

	for _, result := range searchResults.Results {
		matchScore := 90

		releaseDate, err := time.Parse("2006-01-02", result.FirstAirDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", result.Name)

			releaseDate = time.Time{}
		}

		searchTitle := yearRegex.ReplaceAllString(item.Title, "")
		foundTitle := yearRegex.ReplaceAllString(result.Name, "")

		searchTitle = prefixRegex.ReplaceAllString(searchTitle, "")
		foundTitle = prefixRegex.ReplaceAllString(foundTitle, "")

		searchTitle = strings.ToLower(searchTitle)
		foundTitle = strings.ToLower(foundTitle)

		searchTitle = utils.CleanSortTitle(searchTitle)
		foundTitle = utils.CleanSortTitle(foundTitle)

		distance := levenshtein.ComputeDistance(searchTitle, foundTitle)

		matchScore -= int(math.Abs(float64(distance)))

		// If the show doesn't have a release data, adjust the score.
		// The reasoning is that the show probably hasn't aired yet if it doesn't have a date,
		// so the user shouldn't have episodes for it.
		if releaseDate.IsZero() {
			matchScore -= 5
		} else {
			matchScore += 5
		}

		// If we have both dates, compare them to adjust the score. The further apart they are, the less likely it's a match.
		if !item.ReleaseDate.IsZero() && !releaseDate.IsZero() {
			difference := item.ReleaseDate.Year() - releaseDate.Year()

			if difference == 0 {
				matchScore += 10
			} else {
				matchScore -= 5 * difference
			}
		}

		// At this point, if the score is too low, drop the result
		if matchScore < 0 {
			continue
		}

		seriesResult := sdk.TVShow{
			ItemInfo: &sdk.ItemInfo{
				Title:         result.Name,
				OriginalTitle: result.OriginalName,
				ReleaseDate:   releaseDate,
				MatchScore:    uint(matchScore),
			},
			Popularity: result.Popularity,
			TmdbID:     result.ID,
		}

		results = append(results, seriesResult)
	}

	// Sort results by MatchScore, so the first item is the highest match
	sort.Slice(results, func(i, j int) bool {
		return results[i].(sdk.TVShow).MatchScore > results[j].(sdk.TVShow).MatchScore //nolint:forcetypeassert
	})

	return results, nil
}

func GetMetadata(item database.ItemMetadata) (database.ItemMetadata, error) {
	// Series
	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.Title)

		return database.ItemMetadata{}, err
	}

	if len(results) == 0 {
		return database.ItemMetadata{}, nil
	}

	resultShow := results[0]

	if media, ok := resultShow.(sdk.TVShow); ok {
		seriesData, err := tmdbAPI.GetTvInfo(media.TmdbID, map[string]string{})
		if err != nil {
			log.Err(err).Msgf("failed to fetch information for series \"%s\": %w", item.Title, err)
		}

		releaseDate, err := time.Parse("2006-01-02", seriesData.FirstAirDate)
		if err != nil {
			log.Err(err).Msgf("Failed to parse release date for series \"%s\"", item.Title)

			releaseDate = time.Time{}
		}

		languageTag, err := language.Parse(seriesData.OriginalLanguage)
		if err != nil {
			log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", item.Title)

			languageTag = language.Und
		}

		var artHash string

		if seriesData.BackdropPath != "" {
			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", seriesData.BackdropPath)

			artHash, err = helpers.SaveExternalImageToCache(artPath)
			if err != nil {
				log.Err(err).Msgf("Failed to download backdrop for series \"%s\"", item.Title)
			}
		}

		var posterHash string

		if seriesData.PosterPath != "" {
			posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", seriesData.PosterPath)

			posterHash, err = helpers.SaveExternalImageToCache(posterPath)
			if err != nil {
				log.Err(err).Msgf("failed to download poster for series \"%s\"", item.Title)
			}
		}

		return database.ItemMetadata{
			Title:            seriesData.Name,
			OriginalTitle:    seriesData.OriginalName,
			SortTitle:        utils.CleanSortTitle(seriesData.Name),
			ReleaseDate:      releaseDate,
			Summary:          seriesData.Overview,
			OriginalLanguage: languageTag.String(),
			Thumb:            posterHash,
			Art:              artHash,
		}, nil
	}

	return database.ItemMetadata{}, errors.New("unsupported item type")
}
