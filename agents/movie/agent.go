package movie

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/metadata"
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

func GetIdentifier() string {
	return "tv.meteorae.agents.movie"
}

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

	searchResults, err := tmdbAPI.SearchMovie(utils.RemoveUnwantedCharacters(item.Title), options)
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

	return []sdk.Item{}, errNoResultsFound
}

func GetMetadata(item database.ItemMetadata) (sdk.Item, error) {
	log.Debug().Uint("item_id", item.ID).Str("title", item.Title).Msgf("Getting metadata for movie")

	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to get movie results %s", item.Title)

		return nil, err
	}

	if len(results) == 0 {
		return nil, errNoResultsFound
	}

	resultMovie := results[0]

	// Get the TMDb ID
	var tmdbID int

	for _, identifier := range resultMovie.GetIdentifiers() {
		if identifier.IdentifierType == sdk.TmdbIdentifier {
			parsedID, err := strconv.ParseInt(identifier.Identifier, 10, 32)
			if err != nil {
				log.Err(err).Msgf("Failed to parse TMDb ID %s", identifier.Identifier)

				return nil, err
			}

			tmdbID = int(parsedID)

			break
		}
	}

	if _, ok := resultMovie.(sdk.Movie); ok && tmdbID != 0 {
		// Get the full movie details
		movieData, err := tmdbAPI.GetMovieInfo(tmdbID, map[string]string{})
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

		var (
			artHash string
			artPath string
		)

		if movieData.BackdropPath != "" {
			artPath = fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

			artHash, err = helpers.SaveExternalImageToCache(artPath, GetIdentifier(), item, "art")
			if err != nil {
				log.Err(err).Msgf("Failed to download backdrop for movie \"%s\"", item.Title)
			}
		}

		var (
			posterHash string
			posterPath string
		)

		if movieData.PosterPath != "" {
			posterPath = fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

			posterHash, err = helpers.SaveExternalImageToCache(posterPath, GetIdentifier(), item, "poster")
			if err != nil {
				log.Err(err).Msgf("failed to download poster for movie \"%s\"", item.Title)
			}
		}

		return sdk.Movie{
			ItemInfo: &sdk.ItemInfo{
				Title:         movieData.Title,
				OriginalTitle: movieData.OriginalTitle,
				ReleaseDate:   releaseDate,
				Language:      languageTag.String(),
				UUID:          item.UUID,
				// TODO: Return all available images here.
				Thumb: sdk.Posters{
					Items: []sdk.ItemImage{
						{
							External:  true,
							Provider:  GetIdentifier(),
							Media:     metadata.GetURIForAgent(GetIdentifier(), posterHash),
							URL:       posterPath,
							SortOrder: 0,
						},
					},
				},
				// TODO: Return all available images here.
				Art: sdk.Art{
					Items: []sdk.ItemImage{
						{
							External:  true,
							Provider:  GetIdentifier(),
							Media:     metadata.GetURIForAgent(GetIdentifier(), artHash),
							URL:       artPath,
							SortOrder: 0,
						},
					},
				},
			},
		}, nil
	}

	return sdk.Movie{}, errors.New("unsupported item type")
}
