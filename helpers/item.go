package helpers

import (
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/models"
)

func GetItemFromItemMetadata(itemMetadata database.ItemMetadata) model.Item {
	var item model.Item

	switch itemMetadata.Type {
	case database.MovieItem:
		item = models.NewMovieFromItemMetadata(itemMetadata)
	case database.CollectionItem,
		database.ImageAlbumItem,
		database.ImageItem,
		database.MusicAlbumItem,
		database.MusicMediumItem,
		database.MusicTrackItem,
		database.PersonItem,
		database.TVEpisodeItem,
		database.TVSeasonItem,
		database.TVShowItem:
		return nil
	}

	return item
}
