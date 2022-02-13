package helpers

import (
	"fmt"
	"strconv"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/graph/model"
)

func GetItemFromItemMetadata(itemMetadata *database.ItemMetadata) *model.Item {
	itemID := strconv.FormatUint(itemMetadata.ID, 10) //nolint:gomnd

	thumbURL := ""
	if itemMetadata.Thumb != "" {
		thumbURL = fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", itemMetadata.ID)
	}

	artURL := ""
	if itemMetadata.Art != "" {
		artURL = fmt.Sprintf("/image/transcode?url=/metadata/%d/art", itemMetadata.ID)
	}

	var item model.Item

	switch itemMetadata.Type {
	case database.MovieItem:
		isoReleaseDate := itemMetadata.ReleaseDate.Format("2006-01-02")

		item = model.Movie{
			ID:          itemID,
			Title:       itemMetadata.Title,
			ReleaseDate: &isoReleaseDate,
			Summary:     &itemMetadata.Summary,
			Thumb:       &thumbURL,
			Art:         &artURL,
			Library:     &itemMetadata.Library,
			CreatedAt:   itemMetadata.CreatedAt,
			UpdatedAt:   itemMetadata.UpdatedAt,
		}
	case database.ImageAlbumItem:
		item = model.ImageAlbum{
			ID:        itemID,
			Title:     itemMetadata.Title,
			Summary:   &itemMetadata.Summary,
			Thumb:     &thumbURL,
			Art:       &artURL,
			Library:   &itemMetadata.Library,
			CreatedAt: itemMetadata.CreatedAt,
			UpdatedAt: itemMetadata.UpdatedAt,
		}
	case database.ImageItem:
		item = model.Image{
			ID:        itemID,
			Title:     itemMetadata.Title,
			Summary:   &itemMetadata.Summary,
			Thumb:     &thumbURL,
			Art:       &artURL,
			Library:   &itemMetadata.Library,
			CreatedAt: itemMetadata.CreatedAt,
			UpdatedAt: itemMetadata.UpdatedAt,
		}
	}

	return &item
}
