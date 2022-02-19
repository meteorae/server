package database

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
)

type ItemType uint

const (
	MovieItem ItemType = iota
	MusicAlbumItem
	MusicMediumItem
	MusicTrackItem
	TVShowItem
	TVSeasonItem
	TVEpisodeItem
	AnimeMovieItem
	AnimeShowItem
	AnimeSeasonItem
	AnimeEpisodeItem
	ImageItem
	ImageAlbumItem
	PersonItem
	GroupItem
	CollectionItem
)

type ItemMetadata struct {
	ID            uint64   `gorm:"primary_key" json:"id"`
	Title         string   `gorm:"type:VARCHAR(255)" json:"title"`
	SortTitle     string   `gorm:"type:VARCHAR(255) COLLATE NOCASE" json:"sortTitle"`
	OriginalTitle string   `gorm:"type:VARCHAR(255)" json:"originalTitle"`
	Tagline       string   `gorm:"type:VARCHAR(255)" json:"tagline"`
	Summary       string   `json:"summary"`
	Type          ItemType `gorm:"not null;type:INT" json:"type"`
	// ExternalID []ExternalIdentifier
	ReleaseDate      time.Time      `json:"releaseDate"`
	Popularity       float32        `json:"popularity"`
	ParentID         uint64         `json:"parentId"`
	Index            int64          `json:"index"`
	AbsoluteIndex    int64          `json:"absoluteIndex"`
	Duration         int64          `json:"duration"`
	OriginalLanguage string         `json:"originalLanguage"`
	Thumb            string         `json:"thumb"`
	Art              string         `json:"art"`
	ExtraInfo        datatypes.JSON `json:"extraInfo"`
	MediaPart        MediaPart      `json:"mediaPart"`
	LibraryID        uint64
	Library          Library   `gorm:"not null" json:"library"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	DeleteAt         time.Time `json:"deleteAt"`
}

type MovieExtraInfo struct {
	Type string `json:"type"`
}

// Returns the requested fields from the specified item.
func GetItemByID(id string) (*ItemMetadata, error) {
	// TODO: Figure out a way to only request specific fields for this
	var item ItemMetadata

	if result := db.First(&item, id); result.Error != nil {
		return nil, result.Error
	}

	return &item, nil
}

// Returns all the top-level items from the specified library.
func GetItemsFromLibrary(libraryID string, limit, offset *int64) ([]*ItemMetadata, error) {
	var items []*ItemMetadata

	result := db.
		Limit(int(*limit)).
		Offset(int(*offset)).
		Where("library_id = ? AND parent_id = 0", libraryID).
		Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func GetItemsCountFromLibrary(libraryID string) (*int64, error) {
	var count int64

	result := db.Model(&ItemMetadata{}).Where("library_id = ? AND parent_id = 0", libraryID).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}

	return &count, nil
}

// Returns all the children of a given item.
func GetChildrenFromItem(id string, limit, offset *int64) ([]*ItemMetadata, error) {
	var children []*ItemMetadata

	result := db.
		Limit(int(*limit)).
		Offset(int(*offset)).
		Where("parent_id = ?", id).
		Find(&children)
	if result.Error != nil {
		return nil, result.Error
	}

	return children, nil
}

// Returns the number of children for a given item.
func GetChildrenCountFromItem(id string) (*int64, error) {
	var count int64

	result := db.Model(&ItemMetadata{}).Where("parent_id = ?", id).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}

	return &count, nil
}

func GetLatestItemsFromLibrary(libraryID uint64, limit int) ([]*ItemMetadata, error) {
	var items []*ItemMetadata

	itemsResult := db.
		Limit(limit).
		Where("library_id = ? AND parent_id = 0", libraryID).
		Order("created_at desc").
		Find(&items)
	if itemsResult.Error != nil {
		return nil, fmt.Errorf("failed to get items: %w", itemsResult.Error)
	}

	return items, nil
}

func CreateMovie(movieInfo *ItemMetadata) error {
	if result := db.Create(movieInfo); result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateMovie(movieInfo *ItemMetadata) error {
	if result := db.Save(movieInfo); result.Error != nil {
		return result.Error
	}

	return nil
}

func CreateImage(imageInfo *ItemMetadata) error {
	if result := db.Create(imageInfo); result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateImage(imageInfo *ItemMetadata) error {
	if result := db.Save(imageInfo); result.Error != nil {
		return result.Error
	}

	return nil
}

func GetImageAlbum(id uint64) (*ItemMetadata, error) {
	var imageAlbum ItemMetadata

	result := db.First(&imageAlbum, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &imageAlbum, nil
}

func GetImageAlbumByPath(path string) (*ItemMetadata, error) {
	var imageAlbumPart MediaPart

	var imageAlbum ItemMetadata

	result := db.Where("file_path = ?", path).First(&imageAlbumPart)
	if result.Error != nil {
		return nil, result.Error
	}

	result = db.Where("id = ?", imageAlbumPart.ID).First(&imageAlbum)
	if result.Error != nil {
		return nil, result.Error
	}

	return &imageAlbum, nil
}

func CreateImageAlbum(imageAlbumInfo *ItemMetadata) error {
	if result := db.Create(imageAlbumInfo); result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateImageAlbum(imageAlbumInfo *ItemMetadata) error {
	if result := db.Save(imageAlbumInfo); result.Error != nil {
		return result.Error
	}

	return nil
}
