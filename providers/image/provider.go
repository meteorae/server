package image

import (
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
)

func GetInformation(item *database.ItemMetadata, library database.Library) error {
	thumb, err := helpers.SaveLocalImageToCache(item.MediaPart.FilePath)
	if err != nil {
		return fmt.Errorf("could not save image to local cache %s: %w", item.MediaPart.FilePath, err)
	}

	item.Thumb = thumb

	err = database.UpdateImage(item)
	if err != nil {
		return fmt.Errorf("could not get information for image %s: %w", item.MediaPart.FilePath, err)
	}

	return nil
}
