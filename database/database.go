package database

import (
	"fmt"

	"github.com/adrg/xdg"
	"github.com/ostafen/clover"
	"gorm.io/gorm"
)

var db *clover.DB //nolint:varnamelen

type CollectionName string

func (c CollectionName) String() string {
	return string(c)
}

const (
	UserCollection               CollectionName = "User"
	ExternalIdentifierCollection CollectionName = "ExternalIdentifier"
	LibraryCollection            CollectionName = "Library"
	LibraryLocationCollection    CollectionName = "LibraryLocation"
	MediaPartCollection          CollectionName = "MediaPart"
	ItemCollection               CollectionName = "Item"
	MediaStreamCollection        CollectionName = "MediaStream"
)

// Creates a new database file.
func NewDatabase() error {
	databaseLocation, dataFileErr := xdg.DataFile("meteorae/database")
	if dataFileErr != nil {
		return fmt.Errorf("could not get path for database: %w", dataFileErr)
	}

	db, err := clover.Open(databaseLocation)
	defer db.Close()

	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	return nil
}

// Creates the database structure.
func initSchema(transaction *gorm.DB) error {
	if err := db.CreateCollection(string(UserCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(ExternalIdentifierCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(LibraryCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(LibraryLocationCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(MediaPartCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(ItemCollection)); err != nil {
		return err
	}

	if err := db.CreateCollection(string(MediaStreamCollection)); err != nil {
		return err
	}

	return nil
}
