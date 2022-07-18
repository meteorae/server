package models

import (
	"fmt"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/rs/zerolog/log"
)

type MusicAlbum struct {
	*MetadataModel
	Title       string
	SortTitle   string
	Artist      Person
	ReleaseDate time.Time
	Mediums     []Medium
	Thumb       string
	Art         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (a MusicAlbum) String() string {
	return a.Title
}

func (a MusicAlbum) ToItemMetadata() database.ItemMetadata {
	return database.ItemMetadata{
		ID:          a.ID,
		Title:       a.Title,
		SortTitle:   a.SortTitle,
		ReleaseDate: a.ReleaseDate,
		Thumb:       a.Thumb,
		Art:         a.Art,
		ParentID:    a.Artist.ID,
		Type:        database.MusicAlbumItem,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		DeletedAt:   a.DeletedAt,
	}
}

// NewTrackFromItemMetadata creates a Artist from a database.ItemMetadata.
// This should only be used when fetching a track from the database.
func NewMusicAlbumFromItemMetadata(m database.ItemMetadata) MusicAlbum {
	thumbURL := ""
	if m.Thumb != "" {
		thumbURL = fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", m.ID)
	}

	artURL := ""
	if m.Art != "" {
		artURL = fmt.Sprintf("/image/transcode?url=/metadata/%d/art", m.ID)
	}

	artist, err := database.GetItemByID(m.ParentID)
	if err != nil {
		log.Err(err).Msg("Failed to get artist for album")
	}

	return MusicAlbum{
		MetadataModel: &MetadataModel{
			ID: m.ID,
		},
		Title:       m.Title,
		SortTitle:   m.SortTitle,
		Artist:      NewPersonFromItemMetadata(artist),
		ReleaseDate: m.ReleaseDate,
		Thumb:       thumbURL,
		Art:         artURL,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

type Medium struct {
	*MetadataModel
	Title     string
	SortTitle string
	Sequence  int64
	Tracks    []Track
	AlbumID   uint
	Thumb     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (m Medium) String() string {
	if m.Title == "" {
		return fmt.Sprintf("Disc %d", m.Sequence)
	} else {
		return m.Title
	}
}

func (m Medium) ToItemMetadata() database.ItemMetadata {
	return database.ItemMetadata{
		ID:        m.ID,
		Title:     m.Title,
		ParentID:  m.AlbumID,
		Type:      database.MusicMediumItem,
		Sequence:  int(m.Sequence),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

// NewTrackFromItemMetadata creates a Medium from a database.ItemMetadata.
// This should only be used when fetching a medium from the database.
func NewMediumFromItemMetadata(m database.ItemMetadata) (Medium, error) {
	return Medium{
		MetadataModel: &MetadataModel{
			ID: m.ID,
		},
		Title:     m.Title,
		Sequence:  int64(m.Sequence),
		AlbumID:   m.ParentID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}, nil
}

type Track struct {
	*MetadataModel
	Title       string
	SortTitle   string
	Artist      []string
	AlbumArtist string
	AlbumName   string
	AlbumID     uint
	Sequence    int
	MediumID    uint
	DiscIndex   int
	Thumb       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (t Track) String() string {
	return t.Title
}

func (t Track) ToItemMetadata() database.ItemMetadata {
	itemMetadata := database.ItemMetadata{
		Title: t.Title,
		Type:  database.MusicTrackItem,
	}

	if t.Sequence > 0 {
		itemMetadata.Sequence = t.Sequence
	}

	if t.MediumID != 0 {
		itemMetadata.ParentID = t.MediumID
	}

	if !t.CreatedAt.IsZero() {
		itemMetadata.CreatedAt = t.CreatedAt
	}

	if !t.UpdatedAt.IsZero() {
		itemMetadata.UpdatedAt = t.UpdatedAt
	}

	if !t.DeletedAt.IsZero() {
		itemMetadata.DeletedAt = t.DeletedAt
	}

	if len(t.Parts) > 0 {
		itemMetadata.Parts = t.Parts
	}

	if t.Thumb != "" {
		itemMetadata.Thumb = t.Thumb
	}

	return itemMetadata
}

// NewTrackFromItemMetadata creates a Track from a database.ItemMetadata.
// This should only be used when fetching a track from the database.
func NewTrackFromItemMetadata(m database.ItemMetadata) (Track, error) {
	album, err := database.GetItemByID(m.ParentID)
	if err != nil {
		return Track{}, err
	}

	artistRelationships, err := database.GetRelationshipsByRelatedItemIDWithType(m.ID, database.RelationshipArtist)
	if err != nil {
		return Track{}, err
	}

	artistIds := make([]uint, len(artistRelationships))
	for _, artistRelationship := range artistRelationships {
		artistIds = append(artistIds, artistRelationship.ItemMetadataID)
	}

	artists, err := database.GetItemsByID(artistIds)
	if err != nil {
		return Track{}, err
	}

	artistsNames := make([]string, len(artists))

	for i, artist := range artists {
		artistsNames[i] = artist.Title
	}

	albumArtist, err := database.GetItemByID(album.ParentID)
	if err != nil {
		return Track{}, err
	}

	return Track{
		MetadataModel: &MetadataModel{
			ID: m.ID,
		},
		Title:       m.Title,
		Artist:      artistsNames,
		AlbumArtist: albumArtist.Title,
		AlbumName:   album.Title,
		AlbumID:     album.ParentID,
		Sequence:    m.Sequence,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}, nil
}
