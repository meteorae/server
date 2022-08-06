package metadata

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func MergeItems(dst, src sdk.Item) sdk.Item {
	if movieSrcItem, ok := src.(sdk.Movie); ok {
		if movieDstItem, ok := dst.(sdk.Movie); ok {
			if movieSrcItem.Title != "" {
				movieDstItem.Title = movieSrcItem.Title
			}

			if movieSrcItem.OriginalTitle != "" {
				movieDstItem.OriginalTitle = movieSrcItem.OriginalTitle
			}

			if movieSrcItem.SortTitle != "" {
				movieDstItem.SortTitle = movieSrcItem.SortTitle
			}

			if !movieSrcItem.ReleaseDate.IsZero() {
				movieDstItem.ReleaseDate = movieSrcItem.ReleaseDate
			}

			// For Thumbs and Art, we want to prepend source to the destination.
			movieDstItem.Thumb.Items = append(movieSrcItem.Thumb.Items, movieDstItem.Thumb.Items...)
			movieDstItem.Art.Items = append(movieSrcItem.Art.Items, movieDstItem.Art.Items...)

			// Then we need to recompute the sort order.
			for i := range movieDstItem.Thumb.Items {
				movieDstItem.Thumb.Items[i].SortOrder = uint(i)
			}

			for i := range movieDstItem.Art.Items {
				movieDstItem.Art.Items[i].SortOrder = uint(i)
			}

			dst = movieDstItem
		}
	}

	return dst
}

func getInfoItemFromItemMetadata(item database.ItemMetadata) sdk.Item {
	switch item.Type {
	case sdk.MovieItem:
		return sdk.Movie{
			ItemInfo: &sdk.ItemInfo{
				ID:            item.ID,
				UUID:          item.UUID,
				Title:         item.Title,
				OriginalTitle: item.OriginalTitle,
				SortTitle:     item.SortTitle,
				ReleaseDate:   item.ReleaseDate,
				CreatedAt:     item.CreatedAt,
				UpdatedAt:     item.UpdatedAt,
				DeletedAt:     item.DeletedAt,
			},
		}
	case sdk.TVShowItem:
		return sdk.TVShow{
			ItemInfo: &sdk.ItemInfo{
				ID:            item.ID,
				UUID:          item.UUID,
				Title:         item.Title,
				OriginalTitle: item.OriginalTitle,
				SortTitle:     item.SortTitle,
				ReleaseDate:   item.ReleaseDate,
				CreatedAt:     item.CreatedAt,
				UpdatedAt:     item.UpdatedAt,
				DeletedAt:     item.DeletedAt,
			},
		}
	case sdk.MusicAlbumItem:
		return sdk.MusicAlbum{
			ItemInfo: &sdk.ItemInfo{
				ID:            item.ID,
				UUID:          item.UUID,
				Title:         item.Title,
				OriginalTitle: item.OriginalTitle,
				SortTitle:     item.SortTitle,
				ReleaseDate:   item.ReleaseDate,
				CreatedAt:     item.CreatedAt,
				UpdatedAt:     item.UpdatedAt,
				DeletedAt:     item.DeletedAt,
			},
		}
	}

	return nil
}

func GetInfoXML(item database.ItemMetadata) (sdk.Item, error) {
	// Remove dashes from the UUID.
	UUID := strings.ReplaceAll(item.UUID.String(), "-", "")
	UUIDPrefix := UUID[:2]

	infoPath, err := xdg.DataFile(
		filepath.Join("meteorae", "metadata", item.Type.String(), UUIDPrefix, UUID, "combined", "info.xml"))
	if err != nil {
		return nil, err
	}

	xmlFile, err := os.Open(infoPath)
	if err != nil {
		log.Err(err).Msg("Failed to open info.xml")

		// If we don't have anything from the XML, just make shit up.
		// We don't add art, thumbs or parts, since it's likely we don't have any of those if we don't have the XML yet.
		return getInfoItemFromItemMetadata(item), nil
	}

	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read info.xml: %w", err)
	}

	switch item.Type {
	case sdk.MovieItem:
		var movie sdk.Movie

		xml.Unmarshal(byteValue, &movie)

		movie.ID = item.ID
		movie.CreatedAt = item.CreatedAt
		movie.UpdatedAt = item.UpdatedAt
		movie.DeletedAt = item.DeletedAt

		return &movie, nil
	case sdk.TVShowItem:
		var tvShow sdk.TVShow

		xml.Unmarshal(byteValue, &tvShow)

		tvShow.ID = item.ID
		tvShow.CreatedAt = item.CreatedAt
		tvShow.UpdatedAt = item.UpdatedAt
		tvShow.DeletedAt = item.DeletedAt

		return &tvShow, nil
	case sdk.MusicAlbumItem:
		var musicAlbum sdk.MusicAlbum

		xml.Unmarshal(byteValue, &musicAlbum)

		musicAlbum.ID = item.ID
		musicAlbum.CreatedAt = item.CreatedAt
		musicAlbum.UpdatedAt = item.UpdatedAt
		musicAlbum.DeletedAt = item.DeletedAt

		return &musicAlbum, nil
	}

	return nil, fmt.Errorf("unsupported item type: %s", item.Type)
}

func SetItemImages(item sdk.Item, itemType sdk.ItemType) error {
	// Get the item's metadata directory.
	itemUUID := strings.ReplaceAll(item.GetUUID().String(), "-", "")
	itemUUIDPrefix := itemUUID[:2]

	if len(item.GetThumbs()) > 0 {
		itemThumb := item.GetThumbs()[0].Media

		agent, hash := GetURIComponents(itemThumb)

		itemThumb = strings.TrimPrefix(itemThumb, "metadata://")

		combinedThumbPath, err := xdg.DataFile(
			filepath.Join("meteorae", "metadata", itemType.String(), itemUUIDPrefix, itemUUID, "combined", "thumb", itemThumb))
		if err != nil {
			return fmt.Errorf("failed to get metadata directory: %w", err)
		}

		itemThumbPath := GetFilepathForAgentAndHash(agent, hash, item.GetUUID().String(), itemType, "thumb")

		err = os.Symlink(itemThumbPath, combinedThumbPath)
		if err != nil {
			return fmt.Errorf("failed to symlink thumb: %w", err)
		}
	}

	if len(item.GetArt()) > 0 {
		itemArt := item.GetArt()[0].Media

		agent, hash := GetURIComponents(itemArt)

		itemArt = strings.TrimPrefix(itemArt, "metadata://")

		combinedArtPath, err := xdg.DataFile(
			filepath.Join("meteorae", "metadata", itemType.String(), itemUUIDPrefix, itemUUID, "combined", "art", itemArt))
		if err != nil {
			return fmt.Errorf("failed to get metadata directory: %w", err)
		}

		itemArtPath := GetFilepathForAgentAndHash(agent, hash, item.GetUUID().String(), itemType, "art")

		err = os.Symlink(itemArtPath, combinedArtPath)
		if err != nil {
			return fmt.Errorf("failed to symlink thumb: %w", err)
		}
	}

	return nil
}

func SaveMetadataToXML(item sdk.Item, itemType sdk.ItemType, agentIdentifier string) error {
	log.Debug().
		Uint("id", item.GetID()).
		Str("title", item.GetTitle()).
		Str("uuid", item.GetUUID().String()).
		Str("identifier", agentIdentifier).
		Msg("Saving metadata info.xml")

	// Get the item's metadata directory.
	itemUUID := strings.ReplaceAll(item.GetUUID().String(), "-", "")
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
