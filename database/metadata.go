package database

import (
	"fmt"
	"time"

	"github.com/ostafen/clover"
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
	Id               string               `clover:"_id"` //nolint:tagliatelle
	Title            string               `clover:"title"`
	SortTitle        string               `clover:"sortTitle"`
	OriginalTitle    string               `clover:"originalTitle"`
	Tagline          string               `clover:"tagline"`
	Summary          string               `clover:"summary"`
	Type             ItemType             `clover:"type"`
	Externald        []ExternalIdentifier `clover:"externalId"`
	ReleaseDate      time.Time            `clover:"releaseDate"`
	Popularity       float32              `clover:"popularity"`
	ParentId         string               `clover:"parentId"`
	Index            int64                `clover:"index"`
	AbsoluteIndex    int64                `clover:"absoluteIndex"`
	Duration         int64                `clover:"duration"`
	OriginalLanguage string               `clover:"originalLanguage"`
	Thumb            string               `clover:"thumb"`
	Art              string               `clover:"art"`
	Path             string               `clover:"path"`
	LibraryID        string               `clover:"libraryId"`
	Library          Library              `clover:"library"`
	CreatedAt        time.Time            `clover:"createdAt"`
	UpdatedAt        time.Time            `clover:"updatedAt"`
	DeleteAt         time.Time            `clover:"deleteAt"`
}

type ExternalIdentifier struct {
	Type  IdentifierType `clover:"type"`
	Value string         `clover:"value"`
}

// Returns the requested fields from the specified item.
func GetItemById(id string) (*ItemMetadata, error) {
	var item ItemMetadata

	itemDocument, err := db.Query(ItemCollection.String()).Where(clover.Field("_id").Eq(id)).FindFirst()
	if err != nil {
		return &item, fmt.Errorf("failed to get library: %w", err)
	}

	if itemDocument == nil {
		return &item, fmt.Errorf("library not found")
	}

	itemDocument.Unmarshal(&item)

	return &item, nil
}

// Returns all the top-level items from the specified library.
func GetItemsFromLibrary(libraryId string, limit, offset *int64) ([]*ItemMetadata, error) {
	var items []*ItemMetadata //nolint:prealloc

	var item *ItemMetadata

	docs, err := db.Query(ItemCollection.String()).Where(clover.Field("libraryId").Eq(libraryId).
		And(clover.Field("parentId").Eq(0))).Skip(int(*offset)).Limit(int(*limit)).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(item)
		items = append(items, item)
	}

	return items, nil
}

func GetItemsCountFromLibrary(libraryId string) (int64, error) {
	var count int

	count, err := db.Query(ItemCollection.String()).Where(clover.Field("libraryId").Eq(libraryId).
		And(clover.Field("parentId").Eq(0))).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}

// Returns all the children of a given item.
func GetChildrenFromItem(id string, limit, offset *int64) ([]*ItemMetadata, error) {
	var children []*ItemMetadata //nolint:prealloc

	var child *ItemMetadata

	docs, err := db.Query(ItemCollection.String()).Where(clover.Field("parent_id").Eq(id)).
		Skip(int(*offset)).Limit(int(*limit)).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(child)
		children = append(children, child)
	}

	return children, nil
}

// Returns the number of children for a given item.
func GetChildrenCountFromItem(id string) (int64, error) {
	var count int

	count, err := db.Query(ItemCollection.String()).Where(clover.Field("parent_id").Eq(id)).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}

func GetLatestItemsFromLibrary(libraryID string, limit int64) ([]*ItemMetadata, error) {
	var items []*ItemMetadata //nolint:prealloc

	var item *ItemMetadata

	docs, err := db.Query(ItemCollection.String()).Where(clover.Field("parentId").Eq("").
		And(clover.Field("libraryId").Eq(libraryID))).Sort(clover.SortOption{"createdAt", -1}).Limit(int(limit)).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(&item)
		items = append(items, item)
	}

	return items, nil
}

func CreateItem(itemInfo *ItemMetadata) (*ItemMetadata, error) {
	itemInfo.CreatedAt = time.Now()
	itemInfo.UpdatedAt = time.Now()

	document := clover.NewDocumentOf(&itemInfo)

	itemId, err := db.InsertOne(ItemCollection.String(), document)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	itemInfo.Id = itemId

	return itemInfo, nil
}

func UpdateItem(id string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()

	query := db.Query(ItemCollection.String()).Where(clover.Field("_id").Eq(id))

	if err := query.Update(updates); err != nil {
		return err
	}

	return nil
}

func GetItemByPath(path string) (*ItemMetadata, error) {
	var item ItemMetadata

	itemDocument, err := db.Query(ItemCollection.String()).Where(clover.Field("path").
		Eq(path)).FindFirst()
	if err != nil {
		return &item, fmt.Errorf("failed to get item: %w", err)
	}

	itemDocument.Unmarshal(&item)

	return &item, nil
}
