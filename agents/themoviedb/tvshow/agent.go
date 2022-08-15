package tvshow

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
	"github.com/ryanbradynd05/go-tmdb"
	"golang.org/x/text/language"
)

var (
	errNoTmdbIDFound       = errors.New("no TMDb ID found")
	errNoResultsFound      = errors.New("no results found")
	errUnsupportedItemType = errors.New("unsupported item type")
	yearRegex              = regexp.MustCompile(`([ ]+\(?[0-9]{4}\)?)`)
	prefixRegex            = regexp.MustCompile(`^[Bb][Bb][Cc] `)
)

func getTmdbClient() *tmdb.TMDb {
	return tmdb.Init(tmdb.Config{
		APIKey:   "c9ae218044f9b20a4fcbba36d543a730",
		Proxies:  nil,
		UseProxy: false,
	})
}

func getIdentifiers(series *tmdb.TV) []sdk.Identifier {
	identifiers := []sdk.Identifier{
		{
			IdentifierType: sdk.TmdbIdentifier,
			Identifier:     fmt.Sprintf("%d", series.ID),
		},
	}

	if series.ExternalIDs != nil {
		if series.ExternalIDs.ImdbID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TvdbIdentifier,
				Identifier:     series.ExternalIDs.ImdbID,
			})
		}

		if series.ExternalIDs.TvdbID != 0 {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TvdbIdentifier,
				Identifier:     fmt.Sprintf("%d", series.ExternalIDs.TvdbID),
			})
		}

		if series.ExternalIDs.FacebookID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.FacebookIdentifier,
				Identifier:     series.ExternalIDs.FacebookID,
			})
		}

		if series.ExternalIDs.TwitterID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TwitterIdentifier,
				Identifier:     series.ExternalIDs.TwitterID,
			})
		}

		if series.ExternalIDs.InstagramID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.InstagramIdentifier,
				Identifier:     series.ExternalIDs.InstagramID,
			})
		}
	}

	return identifiers
}

func getImages(result *tmdb.TV, item sdk.Item) (sdk.Art, sdk.Posters) {
	var (
		filteredPosters []tmdb.TvImage
		filteredArt     []tmdb.TvImage
	)

	if result.Images == nil {
		return sdk.Art{}, sdk.Posters{}
	}

	for _, image := range result.Images.Posters {
		// TODO: Make this configurable
		if image.Iso639_1 == "en" { //nolint:nosnakecase // From external library.
			filteredPosters = append(filteredPosters, image)
		}
	}

	for _, image := range result.Images.Backdrops {
		if image.Iso639_1 == "en" { //nolint:nosnakecase // From external library.
			filteredArt = append(filteredArt, image)
		}
	}

	sort.Slice(filteredPosters, func(i, j int) bool {
		return filteredPosters[i].VoteAverage > filteredPosters[j].VoteAverage
	})

	sort.Slice(filteredArt, func(i, j int) bool {
		return filteredArt[i].VoteAverage > filteredArt[j].VoteAverage
	})

	var (
		moviePosters = sdk.Posters{
			Items: []sdk.ItemImage{},
		}
		movieArt = sdk.Art{
			Items: []sdk.ItemImage{},
		}
	)

	if len(filteredPosters) > 0 {
		for index, poster := range filteredPosters {
			posterPath := fmt.Sprintf("https://image.tmdb.org/t/p/original%s", poster.FilePath)

			posterHash, posterSaveErr := helpers.SaveExternalImageToCache(
				posterPath, "tv.meteorae.agents.fanarttv", item, "thumb")
			if posterSaveErr != nil {
				log.Err(posterSaveErr).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
			}

			moviePosters.Items = append(moviePosters.Items, sdk.ItemImage{
				External:  true,
				Provider:  "tv.meteorae.agents.fanarttv",
				Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", posterHash),
				URL:       poster.FilePath,
				SortOrder: uint(index),
			})
		}
	}

	if len(filteredArt) > 0 {
		for index, art := range filteredArt {
			artPath := fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", art.FilePath)

			artHash, artSaveErr := helpers.SaveExternalImageToCache(artPath, "tv.meteorae.agents.fanarttv", item, "art")
			if artSaveErr != nil {
				log.Err(artSaveErr).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
			}

			movieArt.Items = append(movieArt.Items, sdk.ItemImage{
				External:  true,
				Provider:  "tv.meteorae.agents.fanarttv",
				Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", artHash),
				URL:       art.FilePath,
				SortOrder: uint(index),
			})
		}
	}

	return movieArt, moviePosters
}

func getTmdbID(item sdk.Item) (int, error) {
	for _, identifier := range item.GetIdentifiers() {
		if identifier.IdentifierType == sdk.TmdbIdentifier {
			parsedID, identifierParseErr := strconv.ParseInt(identifier.Identifier, 10, 32)
			if identifierParseErr != nil {
				log.Err(identifierParseErr).Msgf("Failed to parse TMDb ID %s", identifier.Identifier)

				return 0, fmt.Errorf("failed to parse TMDb ID: %w", identifierParseErr)
			}

			return int(parsedID), nil
		}
	}

	return 0, errNoTmdbIDFound
}

func parseSeriesInfo(series *tmdb.TV) (time.Time, string) {
	releaseDate, movieInfoFetchErr := time.Parse("2006-01-02", series.FirstAirDate)
	if movieInfoFetchErr != nil {
		log.Err(movieInfoFetchErr).Msgf("Failed to parse release date for movie \"%s\"", series.Name)

		releaseDate = time.Time{}
	}

	var languageTag string

	languageBase, languageParseErr := language.ParseBase(series.OriginalLanguage)
	if languageParseErr != nil {
		log.Debug().
			Err(languageParseErr).
			Msgf("Failed to parse original language for movie \"%s\", using Undefined", series.Name)

		languageTag = language.Und.String()
	} else {
		languageTag = languageBase.String()
	}

	return releaseDate, languageTag
}

// FIXME: Kingdom (2019) is matching to Kingdom (2017).
func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	tmdbAPI := getTmdbClient()

	options := map[string]string{
		"language":      "en-US", // TODO: Make this configurable
		"include_adult": "false", // TODO: Make this configurable
	}

	if !item.GetReleaseDate().IsZero() {
		options["year"] = fmt.Sprintf("%d", item.GetReleaseDate().Year())
	}

	searchResults, searchErr := tmdbAPI.SearchTv(item.GetTitle(), options)
	if searchErr != nil {
		log.Err(searchErr).Msgf("Failed to search for series %s", item.GetTitle())

		return nil, fmt.Errorf("failed to search for series: %w", searchErr)
	}

	results := make([]sdk.Item, 0, len(searchResults.Results))

	for _, result := range searchResults.Results {
		matchScore := 90

		releaseDate, dateParseErr := time.Parse("2006-01-02", result.FirstAirDate)
		if dateParseErr != nil {
			log.Err(dateParseErr).Msgf("Failed to parse release date for movie \"%s\"", result.Name)

			releaseDate = time.Time{}
		}

		searchTitle := yearRegex.ReplaceAllString(item.GetTitle(), "")
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
		if !item.GetReleaseDate().IsZero() && !releaseDate.IsZero() {
			difference := item.GetReleaseDate().Year() - releaseDate.Year()

			if difference == 0 {
				matchScore += 10
			} else {
				matchScore -= 5 * difference //nolint:gomnd // Weight the difference more if it's a large number.
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
				Identifiers: []sdk.Identifier{
					{
						IdentifierType: sdk.TmdbIdentifier,
						Identifier:     fmt.Sprintf("%d", result.ID),
					},
				},
			},
			Popularity: result.Popularity,
		}

		results = append(results, seriesResult)
	}

	// Sort results by MatchScore, so the first item is the highest match
	sort.Slice(results, func(i, j int) bool {
		return results[i].(sdk.TVShow).MatchScore > //nolint:forcetypeassert // This is always a TV show.
			results[j].(sdk.TVShow).MatchScore //nolint:forcetypeassert // This is always a TV show.
	})

	return results, nil
}

func GetMetadata(item sdk.Item) (sdk.Item, error) {
	tmdbAPI := getTmdbClient()

	log.Debug().
		Str("identifier", "tv.meteorae.agents.tmdb").
		Uint("item_id", item.GetID()).
		Str("title", item.GetTitle()).
		Msgf("Getting metadata for TV show")

	// Series
	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.GetTitle())

		return nil, err
	}

	if len(results) == 0 {
		return nil, errNoResultsFound
	}

	resultShow := results[0]

	// Get the TMDb ID
	tmdbID, getTmdbIDErr := getTmdbID(resultShow)
	if getTmdbIDErr != nil {
		log.Err(getTmdbIDErr).Msgf("Failed to get TMDb ID for %d", resultShow.GetID())

		return nil, getTmdbIDErr
	}

	if seriesItem, ok := resultShow.(sdk.TVShow); ok && tmdbID != 0 {
		seriesData, showInfoFetchErr := tmdbAPI.GetTvInfo(tmdbID, map[string]string{})
		if showInfoFetchErr != nil {
			log.Err(showInfoFetchErr).Msgf("failed to fetch information for series \"%s\"", item.GetTitle())
		}

		if seriesData == nil {
			log.Debug().Msgf("No data found for series \"%s\"", item.GetTitle())

			return nil, errNoResultsFound
		}

		releaseDate, languageTag := parseSeriesInfo(seriesData)

		movieArt, moviePosters := getImages(seriesData, item)

		identifiers := getIdentifiers(seriesData)

		return sdk.TVShow{
			ItemInfo: &sdk.ItemInfo{
				ID:            seriesItem.ID,
				UUID:          seriesItem.UUID,
				Title:         seriesData.Name,
				OriginalTitle: seriesData.OriginalName,
				ReleaseDate:   releaseDate,
				Language:      languageTag,
				Identifiers:   identifiers,
				Thumb:         moviePosters,
				Art:           movieArt,
			},
		}, nil
	}

	return nil, errUnsupportedItemType
}
