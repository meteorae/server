package agents

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/meteorae/meteorae-server/agents/fanart"
	"github.com/meteorae/meteorae-server/agents/themoviedb"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/tasks"
	"github.com/rs/zerolog/log"
)

// Scanners are organised by item type.
var scanners = map[database.LibraryType]map[string][]sdk.Agent{}

func InitAgentsManager() {
	// TODO: Agents should be populated from plugins.

	tmdbMovieAgent := themoviedb.MoviePlugin.GetMovieAgent()
	tmdbTVAgent := themoviedb.TVPlugin.GetTVAgent()

	fanartMovieAgent := fanart.MoviePlugin.GetMovieAgent()

	scanners[database.MovieLibrary] = map[string][]sdk.Agent{
		tmdbMovieAgent.Identifier: {
			*tmdbMovieAgent,
			*fanartMovieAgent,
		},
	}

	scanners[database.TVLibrary] = map[string][]sdk.Agent{
		tmdbTVAgent.Identifier: {
			*tmdbTVAgent,
		},
	}

	/*scanners[database.MusicLibrary] = map[string][]Agent{
		musicAgent.GetIdentifier(): {
			{
				Identifier:           musicAgent.GetIdentifier(),
				Name:                 musicAgent.GetName(),
				GetMetadataFunc:      musicAgent.GetMetadata,
				GetSearchResultsFunc: musicAgent.GetSearchResults,
			},
		},
	}*/
}

func GetAgentNamesForLibraryType(libraryType database.LibraryType) []*models.Agent {
	scannerNames := make([]*models.Agent, 0, len(scanners[libraryType]))

	for identifier, agent := range scanners[libraryType] {
		scannerNames = append(scannerNames, &models.Agent{
			// All components for an agent should have the same name, so take the first one.
			Name:       agent[0].Name,
			Identifier: identifier,
		})
	}

	return scannerNames
}

func GetAgentComponentsByName(name string, libraryType database.LibraryType) []sdk.Agent {
	for agent, components := range scanners[libraryType] {
		if agent == name {
			return components
		}
	}

	return nil
}

func RefreshItemMetadata(item database.ItemMetadata, library database.Library) error {
	// Get the agent for this item.
	components := GetAgentComponentsByName(library.Agent, library.Type)
	if len(components) == 0 {
		return fmt.Errorf("no agents found for item type: %s", item.Type)
	}

	var combinedItem sdk.Item

	// TODO: Respect a configurable order of agents.
	for _, component := range components {
		// Get the metadata.
		updatedItem, err := component.GetMetadataFunc(item.ToItem())
		if err != nil {
			return fmt.Errorf("failed to get metadata for item: %w", err)
		}

		// Do we have images to save?
		if len(updatedItem.GetThumbs()) > 0 {
			for i, thumb := range updatedItem.GetThumbs() {
				// Use the first thumbnail by default.
				if i == 0 {
					item.Thumb = thumb.Media

					break
				}
			}
		}

		if len(updatedItem.GetArt()) > 0 {
			for i, art := range updatedItem.GetArt() {
				// Use the first thumbnail by default.
				if i == 0 {
					item.Art = art.Media

					break
				}
			}
		}

		// Do we have identifiers to save?
		if len(updatedItem.GetIdentifiers()) > 0 {
			for _, identifier := range updatedItem.GetIdentifiers() {
				if len(item.ExternalIdentifiers) == 0 {
					item.ExternalIdentifiers = []database.ExternalIdentifier{
						{
							IdentifierType: identifier.IdentifierType,
							Identifier:     identifier.Identifier,
						},
					}
				} else {
					item.ExternalIdentifiers = append(item.ExternalIdentifiers, database.ExternalIdentifier{
						IdentifierType: identifier.IdentifierType,
						Identifier:     identifier.Identifier,
					})
				}
			}
		}

		// Save the metadata.
		err = metadata.SaveMetadataToXML(updatedItem, item.Type, component.Identifier)
		if err != nil {
			return fmt.Errorf("failed to save metadata for item: %w", err)
		}

		mergo.Merge(item, item, mergo.WithAppendSlice)

		if combinedItem == nil {
			combinedItem = updatedItem
		} else {
			combinedItem = metadata.MergeItems(combinedItem, updatedItem)
		}
	}

	err := item.Update(item)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	// Save the combined metadata.
	err = metadata.SaveMetadataToXML(combinedItem, item.Type, "combined")
	if err != nil {
		return fmt.Errorf("failed to save metadata for item: %w", err)
	}

	// Make symbolic links for the default images.
	err = metadata.SetItemImages(combinedItem, item.Type)
	if err != nil {
		return fmt.Errorf("failed to set item images: %w", err)
	}

	return nil
}

func RefreshLibraryMetadata(library database.Library) error {
	// Figure out which item types we need to start with.
	var itemTypeToRefresh sdk.ItemType

	switch library.Type {
	case database.MovieLibrary:
		itemTypeToRefresh = sdk.MovieItem
	case database.TVLibrary:
		itemTypeToRefresh = sdk.TVShowItem
	case database.MusicLibrary:
		itemTypeToRefresh = sdk.MusicAlbumItem
	}

	// Get all the items for this library.
	items, err := database.GetItemByLibrayAndType(library, itemTypeToRefresh)
	if err != nil {
		return fmt.Errorf("failed to get items for library: %w", err)
	}

	// Refresh the metadata for each item.
	for i := 0; i < len(items); i++ {
		item := items[i]

		err = tasks.MetadataRefreshQueue.Submit(func() {
			err := RefreshItemMetadata(item, library)
			if err != nil {
				log.Err(err).Stack().Msgf("Failed to refresh metadata for item: %s", item.Title)
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to dispatch library scan task")
		}
	}

	return nil
}
