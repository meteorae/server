package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errInvalidLibraryType = errors.New("invalid library type")

type LibraryType string

const (
	MovieLibrary LibraryType = "movie"
	TVLibrary    LibraryType = "tv"
	MusicLibrary LibraryType = "music"
	ImageLibrary LibraryType = "photo"
)

func (l LibraryType) String() string {
	return string(l)
}

func LibraryTypeFromString(input string) (LibraryType, error) {
	var library LibraryType

	err := library.UnmarshalText([]byte(input))
	if err != nil {
		return library, fmt.Errorf("failed to parse library type: %w", err)
	}

	return library, nil
}

func (l *LibraryType) MarshalText() []byte {
	text := []byte(l.String())

	return text
}

func (l *LibraryType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "movie":
		*l = MovieLibrary

		return nil
	case "tv":
		*l = TVLibrary

		return nil
	case "music":
		*l = MusicLibrary

		return nil
	case "photo":
		*l = ImageLibrary

		return nil
	}

	return fmt.Errorf("%w: %s", errInvalidLibraryType, text)
}

type Library struct {
	ID               uint              `gorm:"primary_key" json:"id"`
	Name             string            `json:"name"`
	Type             LibraryType       `json:"type"`
	UUID             uuid.UUID         `json:"uuid"`
	Language         string            `json:"language"`
	LibraryLocations []LibraryLocation `gorm:"not null" json:"libraryLocations"`
	Scanner          string            `json:"scanner"`
	Agent            string            `json:"agent"`
	Settings         string            `json:"settings"`
	Children         []ItemMetadata    `gorm:"foreignKey:ParentID"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	ScannedAt        time.Time         `json:"scannedAt"`

	IsScanning bool `gorm:"-" json:"isScanning"`
}

func (Library) TableName() string {
	return "libraries"
}

func (library *Library) AfterCreate(*gorm.DB) error {
	for _, observer := range SubsciptionsManager.LibraryAddedObservers {
		observer <- library
	}

	return nil
}

func (library *Library) AfterUpdate(*gorm.DB) error {
	for _, observer := range SubsciptionsManager.LibraryUpdatedObservers {
		observer <- library
	}

	return nil
}

type LibraryLocation struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	LibraryID uint64    `gorm:"not null"`
	RootPath  string    `gorm:"not null" json:"rootPath"`
	Available bool      `gorm:"not null" json:"available"`
	ScannedAt time.Time `json:"scannedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func CreateLibrary(
	name, language, typeArg string,
	locations []string,
	scanner, agent string,
) (*Library, []LibraryLocation, error) {
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
		Scanner:          scanner,
		Agent:            agent,
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
