package database

import (
	"fmt"
	"time"

	"github.com/ostafen/clover"
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

// Given an input string, returns a valid LibraryType.
func LibraryTypeFromString(input string) LibraryType {
	var libraryType LibraryType

	libraryType.UnmarshalText([]byte(input))

	return libraryType
}

func (l *LibraryType) MarshalText() []byte {
	text := []byte(l.String())

	return text
}

func (l *LibraryType) UnmarshalText(text []byte) {
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
}

type Library struct {
	Id               uint64            `clover:"_id"`
	Name             string            `clover:"name"`
	Type             LibraryType       `clover:"type"`
	Language         string            `clover:"language"`
	LibraryLocations []LibraryLocation `clover:"libraryLocations"`
	CreatedAt        time.Time         `clover:"createdAt"`
	UpdatedAt        time.Time         `clover:"updatedAt"`
	ScannedAt        time.Time         `clover:"scannedAt"`
}

type LibraryLocation struct {
	Id        uint64    `clover:"_id"`
	LibraryId uint64    `clover:"libraryId"`
	RootPath  string    `clover:"rootPath"`
	Available bool      `clover:"available"`
	ScannedAt time.Time `clover:"scannedAt"`
	CreatedAt time.Time `clover:"createdAt"`
	UpdatedAt time.Time `clover:"updatedAt"`
}

// Creates a new library entry in the database with the specified information.
func CreateLibrary(name, language, typeArg string, locations []string) (*Library, []LibraryLocation, error) {
	var libraryLocations []LibraryLocation //nolint:prealloc
	for _, location := range locations {
		libraryLocations = append(libraryLocations, LibraryLocation{
			RootPath:  location,
			Available: true,
		})
	}

	libraryType := LibraryTypeFromString(typeArg)

	library := Library{
		Name:             name,
		Type:             libraryType,
		Language:         language,
		LibraryLocations: libraryLocations,
	}

	document := clover.NewDocumentOf(&library)

	if _, err := db.InsertOne(LibraryCollection.String(), document); err != nil {
		return nil, nil, fmt.Errorf("failed to create library: %w", err)
	}

	return &library, libraryLocations, nil
}

// Returns the specified library.
func GetLibraryById(id string) (Library, error) {
	var library Library

	libraryDocument, err := db.Query(LibraryCollection.String()).Where(clover.Field("_id").Eq(id)).FindFirst()
	if err != nil {
		return library, fmt.Errorf("failed to get library: %w", err)
	}

	libraryDocument.Unmarshal(&library)

	return library, nil
}

// Returns the requested fields for all libraries.
func GetLibraries() ([]*Library, error) {
	var libraries []*Library

	var library *Library

	docs, err := db.Query(LibraryCollection.String()).FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get libraries: %w", err)
	}

	for _, doc := range docs {
		doc.Unmarshal(library)
		libraries = append(libraries, library)
	}

	return libraries, nil
}

// Returns the total number of libraries.
func GetLibrariesCount() (int64, error) {
	var count int

	count, err := db.Query(LibraryCollection.String()).Count()
	if err != nil {
		return int64(count), err
	}

	return int64(count), nil
}
