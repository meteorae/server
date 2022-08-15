package movie

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

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
)

func getTmdbClient() *tmdb.TMDb {
	return tmdb.Init(tmdb.Config{
		APIKey:   "c9ae218044f9b20a4fcbba36d543a730",
		Proxies:  nil,
		UseProxy: false,
	})
}

func getIdentifiers(movieData *tmdb.Movie) []sdk.Identifier {
	identifiers := []sdk.Identifier{
		{
			IdentifierType: sdk.TmdbIdentifier,
			Identifier:     fmt.Sprintf("%d", movieData.ID),
		},
	}

	if movieData.ExternalIDs != nil {
		if movieData.ExternalIDs.ImdbID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TvdbIdentifier,
				Identifier:     movieData.ExternalIDs.ImdbID,
			})
		}

		if movieData.ExternalIDs.FacebookID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.FacebookIdentifier,
				Identifier:     movieData.ExternalIDs.FacebookID,
			})
		}

		if movieData.ExternalIDs.TwitterID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TwitterIdentifier,
				Identifier:     movieData.ExternalIDs.TwitterID,
			})
		}

		if movieData.ExternalIDs.TwitterID != "" {
			identifiers = append(identifiers, sdk.Identifier{
				IdentifierType: sdk.TwitterIdentifier,
				Identifier:     movieData.ExternalIDs.TwitterID,
			})
		}
	}

	return identifiers
}

func getImages(result tmdb.Movie, item sdk.Item) (sdk.Art, sdk.Posters) {
	var (
		filteredPosters []tmdb.MovieImage
		filteredArt     []tmdb.MovieImage
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

func parseMovieInfo(movie *tmdb.Movie) (time.Time, string) {
	releaseDate, movieInfoFetchErr := time.Parse("2006-01-02", movie.ReleaseDate)
	if movieInfoFetchErr != nil {
		log.Err(movieInfoFetchErr).Msgf("Failed to parse release date for movie \"%s\"", movie.Title)

		releaseDate = time.Time{}
	}

	var languageTag string

	languageBase, languageParseErr := language.ParseBase(movie.OriginalLanguage)
	if languageParseErr != nil {
		log.Debug().
			Err(languageParseErr).
			Msgf("Failed to parse original language for movie \"%s\", using Undefined", movie.Title)

		languageTag = language.Und.String()
	} else {
		languageTag = languageBase.String()
	}

	return releaseDate, languageTag
}

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	tmdbAPI := getTmdbClient()

	options := map[string]string{
		"language":      "en-US", // TODO: Make this configurable
		"include_adult": "false", // TODO: Make this configurable
	}

	if !item.GetReleaseDate().IsZero() {
		options["year"] = fmt.Sprintf("%d", item.GetReleaseDate().Year())
	}

	searchResults, err := tmdbAPI.SearchMovie(utils.RemoveUnwantedCharacters(item.GetTitle()), options)
	if err != nil {
		log.Err(err).Msgf("Failed to search for movie %s", item.GetTitle())

		return nil, fmt.Errorf("failed to search for movie: %w", err)
	}

	results := make([]sdk.Item, 0, len(searchResults.Results))

	for _, result := range searchResults.Results {
		releaseDate, dateParseErr := time.Parse("2006-01-02", result.ReleaseDate)
		if dateParseErr != nil {
			log.Err(dateParseErr).Msgf("Failed to parse release date for movie \"%s\"", result.Title)

			releaseDate = time.Time{}
		}

		results = append(results, sdk.Movie{
			ItemInfo: &sdk.ItemInfo{
				Title:         result.Title,
				OriginalTitle: result.OriginalTitle,
				ReleaseDate:   releaseDate,
				Identifiers: []sdk.Identifier{
					{
						IdentifierType: sdk.TmdbIdentifier,
						Identifier:     fmt.Sprintf("%d", result.ID),
					},
				},
			},
			Popularity: result.Popularity,
		})
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, errNoResultsFound
}

func GetMetadata(item sdk.Item) (sdk.Item, error) {
	tmdbAPI := getTmdbClient()

	if movieItem, ok := item.(sdk.Movie); ok {
		log.Debug().
			Str("identifier", "tv.meteorae.agents.tmdb").
			Uint("item_id", movieItem.ID).
			Str("title", movieItem.Title).
			Msgf("Getting metadata for movie")

		results, err := GetSearchResults(item)
		if err != nil {
			log.Err(err).Msgf("Failed to get movie results %s", movieItem.Title)

			return nil, err
		}

		if len(results) == 0 {
			return nil, errNoResultsFound
		}

		resultMovie := results[0]

		// Get the TMDb ID
		tmdbID, getTmdbIDErr := getTmdbID(resultMovie)
		if getTmdbIDErr != nil {
			log.Err(getTmdbIDErr).Msgf("Failed to get TMDb ID for %d", movieItem.ID)

			return nil, getTmdbIDErr
		}

		if _, movieItemOk := resultMovie.(sdk.Movie); movieItemOk && tmdbID != 0 {
			// Get the full movie details
			movieData, movieInfoFetchErr := tmdbAPI.GetMovieInfo(tmdbID, map[string]string{})
			if movieInfoFetchErr != nil {
				log.Err(movieInfoFetchErr).Msgf("failed to fetch information for movie \"%s\"", movieItem.Title)
			}

			if movieData == nil {
				log.Debug().Msgf("No movie data found for %d", movieItem.ID)

				return nil, errNoResultsFound
			}

			releaseDate, languageTag := parseMovieInfo(movieData)

			movieArt, moviePosters := getImages(*movieData, item)

			identifiers := getIdentifiers(movieData)

			return sdk.Movie{
				ItemInfo: &sdk.ItemInfo{
					ID:            movieItem.ID,
					UUID:          movieItem.UUID,
					Title:         movieData.Title,
					OriginalTitle: movieData.OriginalTitle,
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

	return nil, errUnsupportedItemType
}
