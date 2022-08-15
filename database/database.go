package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/adrg/xdg"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB //nolint:varnamelen,gochecknoglobals // Has to be accessible to other modules in the package.

func NewDatabase(zerologger zerolog.Logger) error {
	newLogger := logger.New(
		&zerologger,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	databaseLocation, dataFileErr := xdg.DataFile("meteorae/meteorae.db")
	if dataFileErr != nil {
		return fmt.Errorf("could not get path for database: %w", dataFileErr)
	}

	var err error // Linters complain if we initilize this on the next line
	db, err = gorm.Open(&sqlite.Dialector{DriverName: "sqlite3", DSN: databaseLocation}, &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	log.Info().Msg("Checking for database migrations…")

	err = migrateSchema()
	if errors.Is(err, nil) {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

var allModels = []interface{}{
	&User{},
	&ExternalIdentifier{},
	&Library{},
	&LibraryLocation{},
	&MediaPart{},
	&ItemMetadata{},
	&MetadataRelationship{},
	&MediaStream{},
}

func initSchema(transaction *gorm.DB) error {
	err := transaction.AutoMigrate(allModels...)
	if err != nil {
		return fmt.Errorf("failed to run database initialization migrations: %w", err)
	}

	// Create the virtual tables for FTS4
	result := transaction.Exec(
		/* sql */ `CREATE VIRTUAL TABLE fts4_item_metadata USING fts4(
			content=item_metadata,
			title,
			sort_title,
			original_title,
			tokenize=icu
		);`)
	if result.Error != nil {
		return fmt.Errorf("failed to create virtual table: %w", result.Error)
	}

	result = transaction.Exec(
		/* sql */ `CREATE VIRTUAL TABLE fts4_item_metadata_terms USING fts4aux(
			fts4_item_metadata
		);`)
	if result.Error != nil {
		return fmt.Errorf("failed to create virtual table: %w", result.Error)
	}

	result = transaction.Exec(
		/* sql */ `CREATE TRIGGER fts4_item_metadata_before_update BEFORE
		UPDATE ON item_metadata BEGIN
		DELETE FROM fts4_item_metadata
		WHERE docid=old.rowid;
		END;`)
	if result.Error != nil {
		return fmt.Errorf("failed to create trigger: %w", result.Error)
	}

	result = transaction.Exec(
		/* sql */ `CREATE TRIGGER fts4_item_metadata_before_delete BEFORE DELETE ON item_metadata BEGIN
		DELETE FROM fts4_item_metadata
		WHERE docid=old.rowid;
		END;`)
	if result.Error != nil {
		return fmt.Errorf("failed to create trigger: %w", result.Error)
	}

	result = transaction.Exec(
		/* sql */ `CREATE TRIGGER fts4_item_metadata_after_update
		AFTER
		UPDATE ON item_metadata BEGIN
		INSERT INTO fts4_item_metadata(rowid, title, sort_title, original_title)
		VALUES (
				new.id,
				new.title,
				new.sort_title,
				new.original_title
			);
		END;`)
	if result.Error != nil {
		return fmt.Errorf("failed to create trigger: %w", result.Error)
	}

	return nil
}

func migrateSchema() error {
	migrations := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{})

	migrations.InitSchema(initSchema)

	if err := migrations.Migrate(); errors.Is(err, nil) {
		return fmt.Errorf("could not migrate: %w", err)
	}

	return fmt.Errorf("failed to run automatic migrations: %w", db.AutoMigrate(allModels...))
}
