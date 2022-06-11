package database

import (
	"fmt"

	"github.com/adrg/xdg"
	"github.com/ostafen/clover"
)

var db *clover.DB //nolint:varnamelen

type CollectionName string

func (c CollectionName) String() string {
	return string(c)
}

const (
	UserCollection      CollectionName = "user"
	LibraryCollection   CollectionName = "library"
	MediaPartCollection CollectionName = "media_part"
	ItemCollection      CollectionName = "item"
)

// Creates a new database file.
func GetDatabase() (*clover.DB, error) {
	databaseLocation, dataFileErr := xdg.DataFile("meteorae/database")
	if dataFileErr != nil {
		return nil, fmt.Errorf("could not get path for database: %w", dataFileErr)
	}

	var err error

	db, err = clover.Open(databaseLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	initSchema(db)

	return db, nil
}

func createCollectionIfNotExist(collection CollectionName, db *clover.DB) error {
	isExists, err := db.HasCollection(collection.String())
	if err != nil {
		return err
	}

	if !isExists {
		err := db.CreateCollection(collection.String())
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates the database structure.
func initSchema(db *clover.DB) error {
	if err := createCollectionIfNotExist(UserCollection, db); err != nil {
		return err
	}

	if err := createCollectionIfNotExist(LibraryCollection, db); err != nil {
		return err
	}

	if err := createCollectionIfNotExist(MediaPartCollection, db); err != nil {
		return err
	}

	if err := createCollectionIfNotExist(ItemCollection, db); err != nil {
		return err
	}

	return nil
}
