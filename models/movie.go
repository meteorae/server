package models

import (
	"time"

	"github.com/meteorae/meteorae-server/database"
)

type Movie struct {
	*MetadataModel
	Title       string
	TitleSort   string
	ReleaseDate time.Time
	Summary     string
	Thumb       string
	Art         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

func (m Movie) String() string {
	if m.ReleaseDate.Year() == 1 {
		return m.Title
	} else {
		return m.Title + " (" + m.ReleaseDate.Format("2006") + ")"
	}
}

func (m Movie) ToItemMetadata() database.ItemMetadata {
	mediaParts := make([]database.MediaPart, len(m.Parts))
	for i, part := range m.Parts {
		mediaParts[i] = database.MediaPart{
			FilePath: part.FilePath,
		}
	}

	return database.ItemMetadata{
		ID:          m.ID,
		Title:       m.Title,
		SortTitle:   m.TitleSort,
		ReleaseDate: m.ReleaseDate,
		Summary:     m.Summary,
		Thumb:       m.Thumb,
		Art:         m.Art,
		Parts:       mediaParts,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

func MovieSliceToItemMetadataSlice(m []Movie) []database.ItemMetadata {
	metadata := make([]database.ItemMetadata, len(m))

	for i, movie := range m {
		metadata[i] = movie.ToItemMetadata()
	}

	return metadata
}

func NewMovieFromItemMetadata(m database.ItemMetadata) *Movie {
	return &Movie{
		MetadataModel: &MetadataModel{
			ID:    m.ID,
			Parts: []MediaPart{},
		},
		Title:       m.Title,
		TitleSort:   m.SortTitle,
		ReleaseDate: m.ReleaseDate,
		Summary:     m.Summary,
		Thumb:       m.Thumb,
		Art:         m.Art,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}
