package models

import (
	"time"
)

type ItemMetadata struct {
	ID            uint64 `gorm:"primary_key" json:"id"`
	Title         string `gorm:"type:VARCHAR(255)" json:"title"`
	SortTitle     string `gorm:"type:VARCHAR(255) COLLATE NOCASE" json:"sortTitle"`
	OriginalTitle string `gorm:"type:VARCHAR(255)" json:"originalTitle"`
	Tagline       string `gorm:"type:VARCHAR(255)" json:"tagline"`
	Summary       string `json:"summary"`
	// ExternalID []ExternalIdentifier
	ReleaseDate      time.Time `json:"releaseDate"`
	Popularity       float32   `json:"popularity"`
	ParentID         uint64    `json:"parentId"`
	Index            int64     `json:"index"`
	AbsoluteIndex    int64     `json:"absoluteIndex"`
	Duration         int64     `json:"duration"`
	OriginalLanguage string    `json:"originalLanguage"`
	Thumb            string    `json:"thumb"`
	Art              string    `json:"art"`
	MediaPart        MediaPart `json:"mediaPart"`
	LibraryID        uint64
	Library          Library   `gorm:"not null" json:"library"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	DeleteAt         time.Time `json:"deleteAt"`
}
