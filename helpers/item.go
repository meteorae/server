package helpers

import (
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/models"
)

func GetItemFromItemMetadata(itemMetadata database.ItemMetadata) models.Item {
	var item models.Item

	switch itemMetadata.Type {
	case database.MovieItem:
		item = models.NewMovieFromItemMetadata(itemMetadata)
	case database.MusicAlbumItem:
		item = models.NewMusicAlbumFromItemMetadata(itemMetadata)
	case database.ImageAlbumItem:
		item = models.NewImageAlbumFromItemMetadata(itemMetadata)
	case database.CollectionItem,
		database.ImageItem,
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
