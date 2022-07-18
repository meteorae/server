package models

import (
	"time"

	"github.com/meteorae/meteorae-server/database"
)

type PhotoAlbum struct {
	*MetadataModel
	Title       string
	SortTitle   string
	ReleaseDate time.Time
	Thumb       string
	Art         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (p PhotoAlbum) String() string {
	return p.Title
}

func (p PhotoAlbum) ToItemMetadata() database.ItemMetadata {
	itemMetadata := database.ItemMetadata{
		Title: p.Title,
		Type:  database.ImageAlbumItem,
	}

	return itemMetadata
}

func NewImageAlbumFromItemMetadata(itemMetadata database.ItemMetadata) *PhotoAlbum {
	return &PhotoAlbum{
		MetadataModel: &MetadataModel{
			ID: itemMetadata.ID,
		},
		Title:       itemMetadata.Title,
		SortTitle:   itemMetadata.SortTitle,
		ReleaseDate: itemMetadata.ReleaseDate,
		Thumb:       itemMetadata.Thumb,
		Art:         itemMetadata.Art,
		CreatedAt:   itemMetadata.CreatedAt,
		UpdatedAt:   itemMetadata.UpdatedAt,
		DeletedAt:   itemMetadata.DeletedAt,
	}
}

type Photo struct {
	*MetadataModel
	Title      string
	PhotoAlbum PhotoAlbum
	Thumb      string
	Art        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
}

func (p Photo) String() string {
	return p.Title
}

func (p Photo) ToItemMetadata() database.ItemMetadata {
	itemMetadata := database.ItemMetadata{
		Title:    p.Title,
		ParentID: p.PhotoAlbum.ID,
		Type:     database.ImageItem,
	}

	if len(p.Parts) > 0 {
		itemMetadata.Parts = p.Parts
	}

	return itemMetadata
}

type VideoClip struct {
	*MetadataModel
	Title      string
	PhotoAlbum PhotoAlbum
	Thumb      string
	Art        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
}

func (c VideoClip) String() string {
	return c.Title
}

func (c VideoClip) ToItemMetadata() database.ItemMetadata {
	itemMetadata := database.ItemMetadata{
		Title: c.Title,
		Type:  database.VideoClipItem,
	}

	if len(c.Parts) > 0 {
		itemMetadata.Parts = c.Parts
	}

	return itemMetadata
}
