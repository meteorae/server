package models

import "time"

type LibraryType string

const (
	MovieLibrary      LibraryType = "movie"
	AnimeMovieLibrary LibraryType = "animeMovie"
	TVLibrary         LibraryType = "tv"
	AnimeTVLibrary    LibraryType = "animeTV"
	MusicLibrary      LibraryType = "music"
)

type Library struct {
	ID               uint64            `gorm:"primary_key" json:"id"`
	Name             string            `json:"name"`
	Type             LibraryType       `json:"type"`
	Language         string            `json:"language"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	ScannedAt        time.Time         `json:"scannedAt"`
	LibraryLocations []LibraryLocation `gorm:"not null" json:"libraryLocations"`
}

type LibraryLocation struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	LibraryID uint64    `gorm:"not null"`
	RootPath  string    `gorm:"not null" json:"rootPath"`
	Available bool      `gorm:"not null" json:"available"`
	ScannedAt time.Time `json:"scannedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
