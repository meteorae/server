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

// Supported image formats for ingestion. Non-supported common formats needing support from libvips are commented out.
// TODO: Check support for RAW formats.
var supportedImageFormats = []string{
	".aiff",
	// ".apng", -- https://github.com/libvips/libvips/issues/2537
	".avif",
	".bmp",
	".gif",
	".jfif",
	".jpeg",
	".jpg",
	".pjpeg",
	".pjp",
	".png",
	".svg",
	".tif",
	".tiff",
	".webp",
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

		if r.isImageFile(file.Name()) {
			return true
		}
	}

	return false
}

func (r Resolver) Resolve(mediaPart *database.MediaPart, library database.Library) error {
	fileName := filepath.Base(mediaPart.Path)

	item := &database.ItemMetadata{
		Title:     fileName,
		Type:      database.ImageAlbumItem,
		LibraryID: library.Id,
		Library:   library,
		Path:      mediaPart.Path,
	}

	item, err := database.CreateItem(item)
	if err != nil {
		return fmt.Errorf("could not resolve image metadata %s: %w", mediaPart.Path, err)
	}

	mediaPart.ItemId = item.Id

	_, err = database.CreateMediaPart(*mediaPart)
	if err != nil {
		return fmt.Errorf("failed to create media part for %s: %w", mediaPart.Path, err)
	}

	return nil
}

func (r Resolver) isImageFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.StringInSlice(ext, supportedImageFormats)
}
