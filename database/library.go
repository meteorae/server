package database

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
	ImageLibrary      LibraryType = "image"
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
	case "image":
		*l = ImageLibrary
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

func CreateLibrary(name, language, typeArg string, locations []string) (*Library, []LibraryLocation, error) {
	var libraryLocations []LibraryLocation //nolint:prealloc
	for _, location := range locations {
		libraryLocations = append(libraryLocations, LibraryLocation{
			RootPath:  location,
			Available: true,
		})
	}

	libraryType, err := LibraryTypeFromString(typeArg)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid library type: %w", err)
	}

	library := Library{
		Name:             name,
		Type:             libraryType,
		Language:         language,
		LibraryLocations: libraryLocations,
	}

	if result := db.Create(&library); result.Error != nil {
		return nil, nil, fmt.Errorf("failed to create library: %w", result.Error)
	}

	return &library, libraryLocations, nil
}

// Returns the requested fields from the specified library.
func GetLibrary(id string) Library {
	var library Library

	db.First(&library, id)

	return library
}

// Returns the requested fields for all libraries.
func GetLibraries() []*Library {
	var libraries []*Library

	db.Find(&libraries)

	return libraries
}

// Returns the total number of libraries.
func GetLibrariesCount() int64 {
	var count int64

	db.Model(&Library{}).Count(&count)

	return count
}
