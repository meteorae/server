package image

import (
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/rs/zerolog/log"
)

func GetInformation(item *database.ItemMetadata, library database.Library) error {
	thumb, err := helpers.SaveLocalImageToCache(item.MediaPart.FilePath)
	if err != nil {
		return fmt.Errorf("could not save image to local cache %s: %w", item.MediaPart.FilePath, err)
	}

	// Figure out if the parent already has a preview thumbnail. If not, we set it to the current item's thumbnail.
	parent, err := database.GetItemById(item.ParentID)
	if err != nil {
		log.Err(err).Msgf("Failed to get image album for path %s: %s", item.MediaPart.FilePath, err)
	}

	if parent.Thumb == "" {
		database.UpdateItem(parent.Id, map[string]interface{}{
			"thumb": thumb,
		})
	}

	err = database.UpdateItem(item.Id, map[string]interface{}{
		"thumb": thumb,
	})
	if err != nil {
		return fmt.Errorf("could not get information for image %s: %w", item.MediaPart.FilePath, err)
	}

	return nil
}
