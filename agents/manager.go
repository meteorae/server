package agents

import (
	"fmt"

	movieAgent "github.com/meteorae/meteorae-server/agents/movie"
	musicAgent "github.com/meteorae/meteorae-server/agents/musicalbum"
	TVShowAgent "github.com/meteorae/meteorae-server/agents/tvshow"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/tasks"
	"github.com/rs/zerolog/log"
)

type (
	GetMetadataFuncType  func(database.ItemMetadata) (database.ItemMetadata, error)
	GetSearchResultsFunc func(database.ItemMetadata) ([]sdk.Item, error)
)

type Agent struct {
	Name                 string
	GetMetadataFunc      GetMetadataFuncType
	GetSearchResultsFunc GetSearchResultsFunc
}

// Scanners are organised by item type.
var scanners = map[database.ItemType][]Agent{}

func InitAgentsManager() {
	scanners[database.MovieItem] = []Agent{
		{
			Name:                 movieAgent.GetName(),
			GetMetadataFunc:      movieAgent.GetMetadata,
			GetSearchResultsFunc: movieAgent.GetSearchResults,
		},
	}
	scanners[database.MusicAlbumItem] = []Agent{
		{
			Name:                 musicAgent.GetName(),
			GetMetadataFunc:      musicAgent.GetMetadata,
			GetSearchResultsFunc: musicAgent.GetSearchResults,
		},
	}
	scanners[database.MusicMediumItem] = []Agent{}
	scanners[database.MusicTrackItem] = []Agent{}
	scanners[database.TVShowItem] = []Agent{
		{
			Name:                 TVShowAgent.GetName(),
			GetMetadataFunc:      TVShowAgent.GetMetadata,
			GetSearchResultsFunc: TVShowAgent.GetSearchResults,
		},
	}
	scanners[database.TVSeasonItem] = []Agent{}
	scanners[database.TVEpisodeItem] = []Agent{}
	scanners[database.ImageItem] = []Agent{}
	scanners[database.ImageAlbumItem] = []Agent{}
	scanners[database.PersonItem] = []Agent{}
	scanners[database.CollectionItem] = []Agent{}
	scanners[database.VideoClipItem] = []Agent{}
}

func GetAgentNamesForLibraryType(libraryType database.LibraryType) []string {
	var itemTypesToCheck []database.ItemType

	// We only check for top-level items here.
	switch libraryType {
	case database.MovieLibrary:
		itemTypesToCheck = []database.ItemType{database.MovieItem}
	case database.TVLibrary:
		itemTypesToCheck = []database.ItemType{database.TVShowItem}
	case database.MusicLibrary:
		itemTypesToCheck = []database.ItemType{database.MusicAlbumItem}
	}

	var scannerNames []string

	// TODO: Should probably remove duplicates here, when/if we add more item types to check.
	for _, itemType := range itemTypesToCheck {
		for _, scanner := range scanners[itemType] {
			scannerNames = append(scannerNames, scanner.Name)
		}
	}

	return scannerNames
}

func GetGetMetadataFuncByName(name string, itemType database.ItemType) GetMetadataFuncType {
	for _, scanner := range scanners[itemType] {
		if scanner.Name == name {
			return scanner.GetMetadataFunc
		}
	}

	return nil
}

func GetGetSearchResultsFunc(name string, itemType database.ItemType) GetSearchResultsFunc {
	for _, scanner := range scanners[itemType] {
		if scanner.Name == name {
			return scanner.GetSearchResultsFunc
		}
	}

	return nil
}

func RefreshItemMetadata(item database.ItemMetadata, library database.Library) error {
	// Get the scanner for this item.
	scanner := GetGetMetadataFuncByName(library.Agent, item.Type)
	if scanner == nil {
		return fmt.Errorf("no scanner found for item type: %s", item.Type)
	}

	// Get the metadata.
	updatedItem, err := scanner(item)
	if err != nil {
		return fmt.Errorf("failed to get metadata for item: %w", err)
	}

	// Save the metadata.
	err = item.Update(updatedItem)
	if err != nil {
		return fmt.Errorf("failed to save metadata for item: %w", err)
	}

	return nil
}

func RefreshLibraryMetadata(library database.Library) error {
	// Figure out which item types we need to start with.
	var itemTypeToRefresh database.ItemType

	switch library.Type {
	case database.MovieLibrary:
		itemTypeToRefresh = database.MovieItem
	case database.TVLibrary:
		itemTypeToRefresh = database.TVShowItem
	case database.MusicLibrary:
		itemTypeToRefresh = database.MusicAlbumItem
	}

	// Get all the items for this library.
	items, err := database.GetItemByLibrayAndType(library, itemTypeToRefresh)
	if err != nil {
		return fmt.Errorf("failed to get items for library: %w", err)
	}

	// Refresh the metadata for each item.
	for _, item := range items {
		err = tasks.MetadataRefreshQueue.Submit(func() {
			item.IsRefreshing = true

			for _, observer := range database.SubsciptionsManager.ItemUpdatedObservers {
				observer <- &item
			}

			err := RefreshItemMetadata(item, library)
			if err != nil {
				log.Error().Msgf("Failed to refresh metadata for item: %s", item.Title)
			}

			item.IsRefreshing = false

			for _, observer := range database.SubsciptionsManager.ItemUpdatedObservers {
				observer <- &item
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to dispatch library scan task")
		}
	}

	return nil
}
