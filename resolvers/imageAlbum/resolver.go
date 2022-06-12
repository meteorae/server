package image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/resolvers/registry"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

func init() {
	registry.Register(imageResolver)
}

var imageResolver registry.Resolver = Resolver{}

type Resolver struct{}

func (r Resolver) GetName() string {
	return "Image Album"
}

func (r Resolver) SupportsLibraryType(library database.Library) bool {
	return library.Type == database.ImageLibrary
}

func (r Resolver) SupportsFileType(filePath string, isDir bool) bool {
	if !isDir {
		return false
	}

	folder, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open directory %s", filePath)

		return false
	}
	defer folder.Close()

	files, err := folder.Readdir(0)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to read directory %s", filePath)

		return false
	}

	// We want to check if any of the files in the directory are images,
	// since photo albums should contain at least one image.
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if utils.IsImageFile(file.Name()) {
			return true
		}
	}

	return false
}

func (r Resolver) Resolve(mediaPart *database.MediaPart, library database.Library) error {
	fileName := filepath.Base(mediaPart.FilePath)

	item := database.ItemMetadata{
		Title:     fileName,
		Type:      database.ImageAlbumItem,
		LibraryID: library.ID,
		Library:   library,
		MediaPart: *mediaPart,
	}

	err := database.CreateImageAlbum(&item)
	if err != nil {
		return fmt.Errorf("could not resolve image metadata %s: %w", mediaPart.FilePath, err)
	}

	return nil
}
