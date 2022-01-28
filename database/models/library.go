package models

import (
	"fmt"
	"time"
)

type LibraryType string

const (
	MovieLibrary      LibraryType = "movie"
	AnimeMovieLibrary LibraryType = "animeMovie"
	TVLibrary         LibraryType = "tv"
	AnimeTVLibrary    LibraryType = "animeTV"
	MusicLibrary      LibraryType = "music"
)

func (l LibraryType) String() string {
	return string(l)
}

func LibraryTypeFromString(input string) (l LibraryType, err error) {
	err = l.UnmarshalText([]byte(input))

	return
}

func (l *LibraryType) MarshalText() (text []byte, err error) {
	text = []byte(l.String())

	return
}

func (l *LibraryType) UnmarshalText(text []byte) (err error) {
	switch string(text) {
	case "movie":
		*l = MovieLibrary
	case "animeMovie":
		*l = AnimeMovieLibrary
	case "tv":
		*l = TVLibrary
	case "animeTV":
		*l = AnimeTVLibrary
	case "music":
		*l = MusicLibrary
	default:
		return fmt.Errorf("unknown library type: %s", string(text))
	}

	return
}

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
