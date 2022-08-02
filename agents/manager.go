package agents

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/imdario/mergo"
	movieAgent "github.com/meteorae/meteorae-server/agents/movie"
	musicAgent "github.com/meteorae/meteorae-server/agents/musicalbum"
	TVShowAgent "github.com/meteorae/meteorae-server/agents/tvshow"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/tasks"
	"github.com/rs/zerolog/log"
)

type (
	GetMetadataFuncType  func(database.ItemMetadata) (sdk.Item, error)
	GetSearchResultsFunc func(database.ItemMetadata) ([]sdk.Item, error)
)

type Agent struct {
	Identifier           string
	Name                 string
	GetMetadataFunc      GetMetadataFuncType
	GetSearchResultsFunc GetSearchResultsFunc
}

// Scanners are organised by item type.
var scanners = map[database.LibraryType][]Agent{}

func InitAgentsManager() {
	scanners[database.MovieLibrary] = []Agent{
		{
			Identifier:           movieAgent.GetIdentifier(),
			Name:                 movieAgent.GetName(),
			GetMetadataFunc:      movieAgent.GetMetadata,
			GetSearchResultsFunc: movieAgent.GetSearchResults,
		},
	}
	scanners[database.MusicLibrary] = []Agent{
		{
			Identifier:           musicAgent.GetIdentifier(),
			Name:                 musicAgent.GetName(),
			GetMetadataFunc:      musicAgent.GetMetadata,
			GetSearchResultsFunc: musicAgent.GetSearchResults,
		},
	}
	scanners[database.TVLibrary] = []Agent{
		{
			Identifier:           TVShowAgent.GetIdentifier(),
			Name:                 TVShowAgent.GetName(),
			GetMetadataFunc:      TVShowAgent.GetMetadata,
			GetSearchResultsFunc: TVShowAgent.GetSearchResults,
		},
	}
}

func GetAgentNamesForLibraryType(libraryType database.LibraryType) []string {
	scannerNames := make([]string, 0, len(scanners[libraryType]))

	for _, scanner := range scanners[libraryType] {
		scannerNames = append(scannerNames, scanner.Name)
	}

	return scannerNames
}

func GetAgentByName(name string, libraryType database.LibraryType) Agent {
	for _, agent := range scanners[libraryType] {
		if agent.Name == name {
			return agent
		}
	}

	return Agent{}
}

func RefreshItemMetadata(item *database.ItemMetadata, library database.Library) error {
	// Get the agent for this item.
	agent := GetAgentByName(library.Agent, library.Type)
	if agent.Identifier == "" {
		return fmt.Errorf("no scanner found for item type: %s", item.Type)
	}

	// Get the metadata.
	updatedItem, err := agent.GetMetadataFunc(*item)
	if err != nil {
		return fmt.Errorf("failed to get metadata for item: %w", err)
	}

	var basicUpdateItem database.ItemMetadata

	// Do we have images to save?
	if len(updatedItem.GetThumbs()) > 0 {
		for _, thumb := range updatedItem.GetThumbs() {
			// Use the first thumbnail by default.
			if thumb.SortOrder == 1 {
				basicUpdateItem.Thumb = thumb.Media

				break
			}
		}
	}

	if len(updatedItem.GetArt()) > 0 {
		for _, art := range updatedItem.GetArt() {
			// Use the first thumbnail by default.
			if art.SortOrder == 1 {
				basicUpdateItem.Art = art.Media

				break
			}
		}
	}

	// Save the metadata.
	err = SaveMetadataToXML(updatedItem, item.Type, agent.Identifier)
	if err != nil {
		return fmt.Errorf("failed to save metadata for item: %w", err)
	}

	err = item.Update(basicUpdateItem)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	mergo.Merge(item, basicUpdateItem, mergo.WithOverride)

	return nil
}

func SaveMetadataToXML(item sdk.Item, itemType database.ItemType, agentIdentifier string) error {
	// Get the item's metadata directory.
	itemUUID := strings.ReplaceAll(item.GetUUID(), "-", "")
	itemUUIDPrefix := itemUUID[:2]

	metadataDir, err := xdg.DataFile(
		filepath.Join("meteorae", "metadata", itemType.String(), itemUUIDPrefix, itemUUID, agentIdentifier, "info.xml"))
	if err != nil {
		return fmt.Errorf("failed to get metadata directory: %w", err)
	}

	var xmlFile []byte

	if media, ok := item.(sdk.Movie); ok {
		xmlFile, err = xml.MarshalIndent(media, "", " ")
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}
	}

	if media, ok := item.(sdk.TVShow); ok {
		xmlFile, err = xml.MarshalIndent(media, "", " ")
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}
	}

	if media, ok := item.(sdk.MusicAlbum); ok {
		xmlFile, err = xml.MarshalIndent(media, "", " ")
		if err != nil {
			return fmt.Errorf("failed to marshal item: %w", err)
		}
	}

	err = os.WriteFile(metadataDir, xmlFile, fs.FileMode(0o644))
	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
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

			err := RefreshItemMetadata(&item, library)
			if err != nil {
				log.Err(err).Msgf("Failed to refresh metadata for item: %s", item.Title)
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
