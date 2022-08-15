package movie

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/meteorae/meteorae-server/agents/fanart/client"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

const timeoutDelay = 5 * time.Second

var (
	errNoTmdbIDFound       = errors.New("no TMDb ID found")
	errNoResultsFound      = errors.New("no results found")
	errUnsupportedItemType = errors.New("unsupported item type")
)

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

func getImages(result *client.MovieResult, item sdk.Item) (sdk.Art, sdk.Posters) {
	sortedPosters := result.Posters
	sort.Slice(sortedPosters, func(i, j int) bool {
		return sortedPosters[i].Likes > sortedPosters[j].Likes
	})

	sortedArt := result.Backgrounds
	sort.Slice(sortedArt, func(i, j int) bool {
		return sortedArt[i].Likes > sortedArt[j].Likes
	})

	var (
		moviePosters = sdk.Posters{
			Items: []sdk.ItemImage{},
		}
		movieArt = sdk.Art{
			Items: []sdk.ItemImage{},
		}
	)

	if len(sortedArt) > 0 {
		for index, art := range sortedArt {
			if art.Lang != "en" {
				continue
			}

			artHash, artSaveErr := helpers.SaveExternalImageToCache(art.URL, "tv.meteorae.agents.fanarttv", item, "art")
			if artSaveErr != nil {
				log.Err(artSaveErr).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
			}

			movieArt.Items = append(movieArt.Items, sdk.ItemImage{
				External:  true,
				Provider:  "tv.meteorae.agents.fanarttv",
				Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", artHash),
				URL:       art.URL,
				SortOrder: uint(index),
			})
		}
	}

	if len(sortedPosters) > 0 {
		for index, poster := range sortedPosters {
			if poster.Lang != "en" {
				continue
			}

			posterHash, posterSaveErr := helpers.SaveExternalImageToCache(
				poster.URL, "tv.meteorae.agents.fanarttv", item, "thumb")
			if posterSaveErr != nil {
				log.Err(posterSaveErr).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
			}

			moviePosters.Items = append(moviePosters.Items, sdk.ItemImage{
				External:  true,
				Provider:  "tv.meteorae.agents.fanarttv",
				Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", posterHash),
				URL:       poster.URL,
				SortOrder: uint(index),
			})
		}
	}

	return movieArt, moviePosters
}

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	return nil, errNoResultsFound
}

func GetMetadata(item sdk.Item) (sdk.Item, error) {
	fanartClient := client.New()

	if movieItem, ok := item.(sdk.Movie); ok {
		log.Debug().
			Str("identifier", "tv.meteorae.agents.fanarttv").
			Uint("item_id", movieItem.ID).
			Str("title", movieItem.Title).
			Msgf("Getting metadata for movie")

		// Get the TMDb ID
		tmdbID, getTmdbIDErr := getTmdbID(item)
		if getTmdbIDErr != nil {
			log.Err(getTmdbIDErr).Msgf("Failed to get TMDb ID for %d", movieItem.ID)

			return nil, getTmdbIDErr
		}

		if tmdbID != 0 {
			ctx, cancel := context.WithTimeout(context.TODO(), timeoutDelay)
			defer cancel()

			movieImages, getMovieErr := fanartClient.GetMovieImages(ctx, fmt.Sprintf("%d", tmdbID))
			if getMovieErr != nil {
				log.Err(getMovieErr).Msgf("Failed to get movie images for %d", tmdbID)

				return nil, fmt.Errorf("failed to get show images: %w", getMovieErr)
			}

			movieArt, moviePosters := getImages(movieImages, item)

			return sdk.Movie{
				ItemInfo: &sdk.ItemInfo{
					ID:    item.GetID(),
					UUID:  item.GetUUID(),
					Thumb: moviePosters,
					Art:   movieArt,
				},
			}, nil
		}

		return nil, errNoTmdbIDFound
	}

	return nil, errUnsupportedItemType
}
