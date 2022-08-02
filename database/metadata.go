package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ItemType uint

func (t ItemType) String() string {
	switch t {
	case MovieItem:
		return "Movie"
	case MusicAlbumItem:
		return "MusicAlbum"
	case MusicMediumItem:
		return "MusicMedium"
	case MusicTrackItem:
		return "MusicTrack"
	case TVSeasonItem:
		return "TVSeason"
	case TVShowItem:
		return "TVShow"
	case TVEpisodeItem:
		return "TVEpisode"
	case ImageItem:
		return "Image"
	case ImageAlbumItem:
		return "ImageAlbum"
	case PersonItem:
		return "Person"
	case CollectionItem:
		return "Collection"
	case VideoClipItem:
		return "VideoClip"
	}

	return "Unknown"
}

const (
	MovieItem ItemType = iota
	MusicAlbumItem
	MusicMediumItem
	MusicTrackItem
	TVShowItem
	TVSeasonItem
	TVEpisodeItem
	ImageItem
	ImageAlbumItem
	PersonItem
	CollectionItem
	VideoClipItem
)

type ItemMetadata struct {
	ID                  uint                 `gorm:"primary_key" json:"id"`
	Title               string               `gorm:"type:VARCHAR(255)" json:"title"`
	SortTitle           string               `gorm:"type:VARCHAR(255) COLLATE NOCASE" json:"sortTitle"`
	OriginalTitle       string               `gorm:"type:VARCHAR(255)" json:"originalTitle"`
	Tagline             string               `gorm:"type:VARCHAR(255)" json:"tagline"`
	Summary             string               `json:"summary"`
	Type                ItemType             `gorm:"not null;type:INT" json:"type"`
	UUID                uuid.UUID            `gorm:"not null;type:UUID" json:"uuid"`
	ExternalIdentifiers []ExternalIdentifier `gorm:"foreignKey:ItemMetadataID"`
	ReleaseDate         time.Time            `json:"releaseDate"`
	EndDate             time.Time            `json:"endDate"`
	Popularity          float32              `json:"popularity"`
	ParentID            uint                 `json:"parentId"`
	Children            []ItemMetadata       `gorm:"foreignkey:ParentID" json:"children"`
	Sequence            int                  `json:"sequence"`
	AbsoluteSequence    int                  `json:"absoluteSequence"`
	Duration            uint                 `json:"duration"`
	OriginalLanguage    string               `json:"originalLanguage"`
	Thumb               string               `json:"thumb"`
	Art                 string               `json:"art"`
	// ExtraInfo        datatypes.JSON `json:"extraInfo"`
	Parts        []MediaPart `gorm:"foreignKey:ItemMetadataID" json:"mediaPart"`
	LibraryID    uint
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    time.Time `json:"deleteAt"`
	IsRefreshing bool      `gorm:"-" json:"isRefreshing"`
	IsAnalyzing  bool      `gorm:"-" json:"isAnalyzing"`
}

func (item *ItemMetadata) AfterCreate(*gorm.DB) error {
	for _, observer := range SubsciptionsManager.ItemAddedObservers {
		observer <- item
	}

	return nil
}

func (item *ItemMetadata) AfterUpdate(*gorm.DB) error {
	for _, observer := range SubsciptionsManager.ItemUpdatedObservers {
		observer <- item
	}

	return nil
}

// Returns the requested fields from the specified item.
func GetItemByID(id uint) (ItemMetadata, error) {
	var item ItemMetadata

	if result := db.Preload("ExternalIdentifiers").First(&item, id); result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

func GetItemByParentWithIndex(parentID uint, index int) (ItemMetadata, error) {
	var item ItemMetadata

	result := db.Where("parent_id = ? AND sequence = ?", parentID, index).First(&item)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

func GetItemsByID(ids []uint) ([]ItemMetadata, error) {
	var items []ItemMetadata

	if result := db.Where("id IN (?)", ids).Find(&items); result.Error != nil {
		return nil, result.Error
	}

	return items, nil
}

func GetItemByUUID(uuid uuid.UUID) (ItemMetadata, error) {
	var item ItemMetadata

	if result := db.Where("uuid = ?", uuid).First(&item); result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

func GetItemByLibrayAndType(library Library, itemType ItemType) ([]ItemMetadata, error) {
	var items []ItemMetadata

	result := db.Where("library_id = ? AND type = ?", library.ID, itemType).Find(&items)
	if result.Error != nil {
		return []ItemMetadata{}, result.Error
	}

	return items, nil
}

func GetItemByUUIDAndType(uuid uuid.UUID, itemType ItemType) (ItemMetadata, error) {
	var item ItemMetadata

	result := db.Where("uuid = ? AND type = ?", uuid, itemType).First(&item)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

func GetItemByTitleAndType(title string, itemType ItemType) (ItemMetadata, error) {
	var item ItemMetadata

	result := db.Where("title = ? AND type = ?", title, itemType).First(&item)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

// Returns all the top-level items from the specified library.
func GetItemsFromLibrary(library Library, limit, offset *int64) ([]*ItemMetadata, error) {
	var items []*ItemMetadata

	switch library.Type {
	case MusicLibrary:
		result := db.
			Limit(int(*limit)).
			Offset(int(*offset)).
			Where("library_id = ? AND type = ?", library.ID, MusicAlbumItem).
			Find(&items)
		if result.Error != nil {
			return nil, result.Error
		}
	default:
		result := db.
			Limit(int(*limit)).
			Offset(int(*offset)).
			Where("library_id = ? AND parent_id = 0", library.ID).
			Find(&items)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return items, nil
}

func GetItemsCountFromLibrary(library Library) (int64, error) {
	var count int64

	switch library.Type {
	case MusicLibrary:
		result := db.Model(&ItemMetadata{}).Where("library_id = ? AND type = ?", library.ID, MusicAlbumItem).Count(&count)
		if result.Error != nil {
			return 0, result.Error
		}
	default:
		result := db.Model(&ItemMetadata{}).Where("library_id = ? AND parent_id = 0", library.ID).Count(&count)
		if result.Error != nil {
			return 0, result.Error
		}
	}

	return count, nil
}

// Returns the first child of a given item with a given type.
func GetChildFromItem(parentItemID uint, itemType ItemType) (ItemMetadata, error) {
	var item ItemMetadata

	result := db.
		Where("parent_id = ? AND type = ?", parentItemID, itemType).
		First(&item)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

// Returns all the children of a given item.
func GetChildrenFromItem(parentItemID uint, limit, offset *int64) ([]ItemMetadata, error) {
	var children []ItemMetadata

	result := db.
		Limit(int(*limit)).
		Offset(int(*offset)).
		Where("parent_id = ?", parentItemID).
		Find(&children)
	if result.Error != nil {
		return nil, result.Error
	}

	return children, nil
}

// Returns the number of children for a given item.
func GetChildrenCountFromItem(id uint) (int64, error) {
	var count int64

	result := db.Model(&ItemMetadata{}).Where("parent_id = ?", id).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

func GetLatestItemsFromLibrary(library Library, limit int) ([]ItemMetadata, error) {
	var items []ItemMetadata

	if library.Type == MovieLibrary || library.Type == ImageLibrary {
		itemsResult := db.
			Limit(limit).
			Where("library_id = ? AND parent_id = 0", library.ID).
			Order("created_at desc").
			Find(&items)
		if itemsResult.Error != nil {
			return nil, fmt.Errorf("failed to get items: %w", itemsResult.Error)
		}
	} else if library.Type == TVLibrary {
		// Eventually replace this by properly grouped episodes
		itemsResult := db.
			Limit(limit).
			Where("library_id = ? AND type = ?", library.ID, TVShowItem).
			Order("updated_at desc").
			Find(&items)
		if itemsResult.Error != nil {
			return nil, fmt.Errorf("failed to get episodes: %w", itemsResult.Error)
		}
	} else if library.Type == MusicLibrary {
		itemsResult := db.
			Limit(limit).
			Where("library_id = ? AND type = ?", library.ID, MusicAlbumItem).
			Order("created_at desc").
			Find(&items)
		if itemsResult.Error != nil {
			return nil, fmt.Errorf("failed to get items: %w", itemsResult.Error)
		}
	}

	return items, nil
}

func CreateItem(item ItemMetadata) (ItemMetadata, error) {
	if result := db.Create(&item); result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

func CreateItemBatch(itemList []ItemMetadata) error {
	if result := db.Create(&itemList); result.Error != nil {
		return result.Error
	}

	return nil
}

func (i *ItemMetadata) Update(update ItemMetadata) error {
	if result := db.Model(&i).Where("id = ?", i.ID).Updates(update); result.Error != nil {
		return result.Error
	}

	if err := mergo.Merge(i, update, mergo.WithOverride); err != nil {
		log.Err(err).Msg("failed to merge updates into item")
	}

	return nil
}

func GetItemByPath(path string) (ItemMetadata, error) {
	var mediaPart MediaPart

	var item ItemMetadata

	result := db.Where("file_path = ?", path).First(&mediaPart)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	result = db.Where("id = ?", mediaPart.ItemMetadataID).First(&item)
	if result.Error != nil {
		return ItemMetadata{}, result.Error
	}

	return item, nil
}

type RelationshipType uint

const (
	RelationshipArtist RelationshipType = iota
	RelationshipAlbumArtist
)

type MetadataRelationship struct {
	ID                    uint      `gorm:"primary_key,type:string" json:"id"`
	ItemMetadataID        uint      `json:"itemMetadataId"`
	RelatedItemMetadataID uint      `json:"relatedItemMetadataId"`
	RelationType          string    `json:"relationType"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (m MetadataRelationship) Create() error {
	if result := db.Create(&m); result.Error != nil {
		return result.Error
	}

	return nil
}

func CreateRelationshipBatch(relationships []MetadataRelationship) error {
	if result := db.Create(&relationships); result.Error != nil {
		return result.Error
	}

	return nil
}

func GetRelationshipsByItemID(itemID uint) ([]MetadataRelationship, error) {
	var relationships []MetadataRelationship

	result := db.Where("item_metadata_id = ?", itemID).Find(&relationships)
	if result.Error != nil {
		return nil, result.Error
	}

	return relationships, nil
}

func GetRelationshipsByRelatedItemId(relatedItemID uint) ([]MetadataRelationship, error) {
	var relationships []MetadataRelationship

	result := db.Where("related_item_metadata_id = ?", relatedItemID).Find(&relationships)
	if result.Error != nil {
		return nil, result.Error
	}

	return relationships, nil
}

func GetRelationshipsByItemIDWithType(
	itemID uint,
	relationType RelationshipType,
) ([]MetadataRelationship, error) {
	var relationships []MetadataRelationship

	result := db.Where("item_metadata_id = ? AND relation_type = ?", itemID, relationType).Find(&relationships)
	if result.Error != nil {
		return nil, result.Error
	}

	return relationships, nil
}

func GetRelationshipsByRelatedItemIDWithType(
	relatedItemID uint,
	relationType RelationshipType,
) ([]MetadataRelationship, error) {
	var relationships []MetadataRelationship

	result := db.Where("related_item_metadata_id = ? AND relation_type = ?", relatedItemID, relationType).
		Find(&relationships)
	if result.Error != nil {
		return nil, result.Error
	}

	return relationships, nil
}
