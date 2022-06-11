package image

import (
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/rs/zerolog/log"
)

func GetInformation(item *database.ItemMetadata, library database.Library) error {
	mediaPart, err := database.GetMediaPartByItemId(item.Id)
	if err != nil {
		return fmt.Errorf("could not find media part for item %s", item.Id)
	}

	thumb, err := helpers.SaveLocalImageToCache(mediaPart.Path)
	if err != nil {
		return fmt.Errorf("could not save image to local cache %s: %w", mediaPart.Path, err)
	}

	// Figure out if the parent already has a preview thumbnail. If not, we set it to the current item's thumbnail.
	parent, err := database.GetItemById(item.ParentId)
	if err != nil {
		log.Err(err).Msgf("Failed to get image album for path %s: %s", mediaPart.Path, err)
	}

	if parent.Thumb == "" {
		err = database.UpdateItem(parent.Id, map[string]interface{}{
			"thumb": thumb,
		})
		if err != nil {
			return fmt.Errorf("could not update image album %s: %w", parent.Id, err)
		}
	}

	err = database.UpdateItem(item.Id, map[string]interface{}{
		"thumb": thumb,
	})
	if err != nil {
		return fmt.Errorf("could not update image %s: %w", mediaPart.Path, err)
	}

	return nil
}
