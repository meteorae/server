package movie

import (
	"errors"
	"fmt"
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
	errNoResultsFound = fmt.Errorf("no results found")
	apiKey            = "c9ae218044f9b20a4fcbba36d543a730" //#nosec
	config            = tmdb.Config{
		APIKey:   apiKey,
		Proxies:  nil,
		UseProxy: false,
	}
)

var tmdbAPI *tmdb.TMDb = tmdb.Init(config)

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
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

func GetMetadata(item sdk.Item) (sdk.Item, error) {
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
				log.Err(err).Msgf("failed to fetch information for movie \"%s\"", movieItem.Title)
			}

			releaseDate, err := time.Parse("2006-01-02", movieData.ReleaseDate)
			if err != nil {
				log.Err(err).Msgf("Failed to parse release date for movie \"%s\"", movieData.Title)

				releaseDate = time.Time{}
			}

			var languageTag string

			languageBase, err := language.ParseBase(movieData.OriginalLanguage)
			if err != nil {
				log.Err(err).Msgf("Failed to parse original language for movie \"%s\", using Undefined", item.GetTitle())

				languageTag = language.Und.String()
			} else {
				languageTag = languageBase.String()
			}

			var (
				artHash string
				artPath string
			)

			if movieData.BackdropPath != "" {
				artPath = fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.BackdropPath)

				artHash, err = helpers.SaveExternalImageToCache(artPath, "tv.meteorae.agents.tmdb", item, "art")
				if err != nil {
					log.Err(err).Msgf("Failed to download backdrop for movie \"%s\"", movieItem.Title)
				}
			}

			var (
				posterHash string
				posterPath string
			)

			if movieData.PosterPath != "" {
				posterPath = fmt.Sprintf("https://image.tmdb.org/t/p/original/%s", movieData.PosterPath)

				posterHash, err = helpers.SaveExternalImageToCache(posterPath, "tv.meteorae.agents.tmdb", item, "thumb")
				if err != nil {
					log.Err(err).Msgf("failed to download poster for movie \"%s\"", movieItem.Title)
				}
			}

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

			return sdk.Movie{
				ItemInfo: &sdk.ItemInfo{
					ID:            movieItem.GetID(),
					Title:         movieData.Title,
					OriginalTitle: movieData.OriginalTitle,
					ReleaseDate:   releaseDate,
					Language:      languageTag,
					UUID:          movieItem.UUID,
					Identifiers:   identifiers,
					// TODO: Return all available images here.
					Thumb: sdk.Posters{
						Items: []sdk.ItemImage{
							{
								External:  true,
								Provider:  "tv.meteorae.agents.tmdb",
								Media:     metadata.GetURIForAgent("tv.meteorae.agents.tmdb", posterHash),
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
								Provider:  "tv.meteorae.agents.tmdb",
								Media:     metadata.GetURIForAgent("tv.meteorae.agents.tmdb", artHash),
								URL:       artPath,
								SortOrder: 0,
							},
						},
					},
				},
			}, nil
		}

		return sdk.Movie{}, errors.New("failed to process search result")
	}

	return sdk.Movie{}, errors.New("unsupported item type")
}
