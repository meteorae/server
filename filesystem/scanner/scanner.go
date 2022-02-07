package scanner

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/resolvers"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

func ScanDirectory(directory string, library database.Library) {
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

		mediaPart, err := database.CreateMediaPart(path, hex.EncodeToString(hash), fileInfo.Size())
		if err != nil {
			return fmt.Errorf("failed to create media part: %w", err)
		}

		// Schedule the file resolution job
		err = ants.Submit(func() {
			log.Debug().Msgf("Scheduling resolution job for %s", mediaPart.FilePath)
			resolvers.ResolveFile(mediaPart, library)
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
