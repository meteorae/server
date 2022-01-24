package scanner

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/resolvers"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ScanDirectory(directory string, database *gorm.DB, libraryType models.LibraryType) {
	err := filepath.WalkDir(directory, func(path string, dirEntry fs.DirEntry, walkErr error) error {
		// TODO: We should probably handle different types differently
		log.Debug().Msgf("Found file: %s", path)

		if helpers.ShouldIgnore(path, dirEntry) {
			if dirEntry.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if dirEntry.IsDir() {
			return nil
		}

		// Hash the file using SHA-1
		hash, err := utils.HashFilePath(path)
		if err != nil {
			return fmt.Errorf("failed to hash file: %w", err)
		}

		fileInfo, err := dirEntry.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		newMediaPart := models.MediaPart{
			FilePath: path,
			Hash:     hex.EncodeToString(hash),
			Size:     fileInfo.Size(),
		}

		// TODO: Maybe batching these into one query is faster?
		// We may run into issues when something is resolved but the create failed and/or conccurency issues
		result := database.Clauses(clause.OnConflict{DoNothing: true}).Create(&newMediaPart)
		// TODO: Check for the actual error type
		if result.Error != nil {
			// If the record exist, we already have it, just skip it to save time
			// TODO: To handle refreshing directories, we'll probably want to
			// get all the existing stuff first, then scan for new files
			return fmt.Errorf("failed to create media part: %w", result.Error)
		}

		// Schedule the file resolution job
		err = ants.Submit(func() {
			log.Debug().Msgf("Scheduling resolution job for %s", newMediaPart.FilePath)
			resolvers.ResolveFile(&newMediaPart, database, libraryType)
		})

		if err != nil {
			return fmt.Errorf("failed to schedule resolution job: %w", err)
		}

		return nil
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to scan %s", directory)
	}
}
