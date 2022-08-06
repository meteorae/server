package movie

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/meteorae/meteorae-server/agents/fanart/client"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

var fanartClient = client.New()

var errNoResultsFound = fmt.Errorf("no results found")

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	return nil, errNoResultsFound
}

func GetMetadata(item sdk.Item) (sdk.Item, error) {
	if movieItem, ok := item.(sdk.Movie); ok {
		log.Debug().
			Str("identifier", "tv.meteorae.agents.fanarttv").
			Uint("item_id", movieItem.ID).
			Str("title", movieItem.Title).
			Msgf("Getting metadata for movie")

		// Get the TMDb ID
		var tmdbID int

		for _, identifier := range item.GetIdentifiers() {
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

		if tmdbID != 0 {
			movieImages, err := fanartClient.GetMovieImages(fmt.Sprintf("%d", tmdbID))
			if err != nil {
				log.Err(err).Msgf("Failed to get movie images for %d", tmdbID)

				return nil, err
			}

			sortedPosters := movieImages.Posters
			sort.Slice(sortedPosters, func(i, j int) bool {
				return sortedPosters[i].Likes > sortedPosters[j].Likes
			})

			sortedArt := movieImages.Backgrounds
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
				for i, art := range sortedArt {
					artHash, err := helpers.SaveExternalImageToCache(art.URL, "tv.meteorae.agents.fanarttv", item, "art")
					if err != nil {
						log.Err(err).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
					}

					movieArt.Items = append(movieArt.Items, sdk.ItemImage{
						External:  true,
						Provider:  "tv.meteorae.agents.fanarttv",
						Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", artHash),
						URL:       art.URL,
						SortOrder: uint(i),
					})
				}
			}

			if len(sortedPosters) > 0 {
				for i, art := range sortedPosters {
					posterHash, err := helpers.SaveExternalImageToCache(art.URL, "tv.meteorae.agents.fanarttv", item, "thumb")
					if err != nil {
						log.Err(err).Msgf("Failed to download backdrop for series \"%s\"", item.GetTitle())
					}

					moviePosters.Items = append(moviePosters.Items, sdk.ItemImage{
						External:  true,
						Provider:  "tv.meteorae.agents.fanarttv",
						Media:     metadata.GetURIForAgent("tv.meteorae.agents.fanarttv", posterHash),
						URL:       art.URL,
						SortOrder: uint(i),
					})
				}
			}

			return sdk.Movie{
				ItemInfo: &sdk.ItemInfo{
					ID:    item.GetID(),
					UUID:  item.GetUUID(),
					Thumb: moviePosters,
					Art:   movieArt,
				},
			}, nil
		}

		return sdk.Movie{}, errors.New("no TMDb ID found")
	}

	return sdk.Movie{}, errors.New("unsupported item type")
}
