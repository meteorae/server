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
	Id            string   `clover:"_id"`
	Title         string   `clover:"title"`
	SortTitle     string   `clover:"sortTitle"`
	OriginalTitle string   `clover:"originalTitle"`
	Tagline       string   `clover:"tagline"`
	Summary       string   `clover:"summary"`
	Type          ItemType `clover:"type"`
	// ExternalID []ExternalIdentifier
	ReleaseDate      time.Time `clover:"releaseDate"`
	Popularity       float32   `clover:"popularity"`
	ParentID         string    `clover:"parentId"`
	Index            int64     `clover:"index"`
	AbsoluteIndex    int64     `clover:"absoluteIndex"`
	Duration         int64     `clover:"duration"`
	OriginalLanguage string    `clover:"originalLanguage"`
	Thumb            string    `clover:"thumb"`
	Art              string    `clover:"art"`
	// ExtraInfo        datatypes.JSON `clover:"extraInfo"`
	MediaPart MediaPart `clover:"mediaPart"`
	LibraryID uint64    `clover:"libraryId"`
	Library   Library   `clover:"library"`
	CreatedAt time.Time `clover:"createdAt"`
	UpdatedAt time.Time `clover:"updatedAt"`
	DeleteAt  time.Time `clover:"deleteAt"`
}

// Returns the requested fields from the specified item.
func GetItemById(id string) (*ItemMetadata, error) {
	var item ItemMetadata

	itemDocument, err := db.Query(ItemCollection.String()).Where(clover.Field("_id").Eq(id)).FindFirst()
	if err != nil {
		return &item, fmt.Errorf("failed to get library: %w", err)
	}

	itemDocument.Unmarshal(&item)

	return &item, nil
}

// Returns all the top-level items from the specified library.
func GetItemsFromLibrary(libraryId string, limit, offset *int64) ([]*ItemMetadata, error) {
	var items []*ItemMetadata

	var item *ItemMetadata

	docs, err := db.Query(ItemCollection.String()).Where(clover.Field("library_id").Eq(libraryId).
		And(clover.Field("parent_id").Eq(0))).Skip(int(*offset)).Limit(int(*limit)).FindAll()
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

	count, err := db.Query(ItemCollection.String()).Where(clover.Field("library_id").Eq(libraryId).
		And(clover.Field("parent_id").Eq(0))).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}

// Returns all the children of a given item.
func GetChildrenFromItem(id string, limit, offset *int64) ([]*ItemMetadata, error) {
	var children []*ItemMetadata

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

func GetLatestItemsFromLibrary(libraryID uint64, limit int) ([]*ItemMetadata, error) {
	var items []*ItemMetadata

	var item *ItemMetadata

	docs, err := db.Query(ItemCollection.String()).Where(clover.Field("parent_id").Eq(0).
		And(clover.Field("library_id").Eq(libraryID))).Sort(clover.SortOption{"created_at", -1}).Limit(limit).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(item)
		items = append(items, item)
	}

	return items, nil
}

func CreateItem(movieInfo *ItemMetadata) error {
	document := clover.NewDocumentOf(&movieInfo)

	if _, err := db.InsertOne(ItemCollection.String(), document); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

func UpdateItem(id string, itemInfo map[string]interface{}) error {
	updates := make(map[string]interface{})

	query := db.Query(MediaPartCollection.String()).Where(clover.Field("_id").Eq(id))

	if err := query.Update(updates); err != nil {
		return err
	}

	return nil
}

func GetItemByPath(path string) (*ItemMetadata, error) {
	var mediaPart MediaPart

	var item ItemMetadata

	partDocument, err := db.Query(ItemCollection.String()).Where(clover.Field("file_path").Eq(path)).FindFirst()
	if err != nil {
		return &item, fmt.Errorf("failed to get media part: %w", err)
	}

	partDocument.Unmarshal(&mediaPart)

	itemDocument, err := db.Query(ItemCollection.String()).Where(clover.Field("itemMetadataId").
		Eq(mediaPart.Id)).FindFirst()
	if err != nil {
		return &item, fmt.Errorf("failed to get item: %w", err)
	}

	itemDocument.Unmarshal(&item)

	return &item, nil
}
