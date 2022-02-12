package image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/providers/image"
	"github.com/meteorae/meteorae-server/resolvers/registry"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/panjf2000/ants/v2"
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
	return "Image"
}

func (r Resolver) SupportsLibraryType(library database.Library) bool {
	return library.Type == database.ImageLibrary
}

func (r Resolver) Resolve(mediaPart *database.MediaPart, library database.Library) error {
	fileInfo, err := os.Stat(mediaPart.FilePath)
	if err != nil {
		return fmt.Errorf("could not stat %s: %w", mediaPart.FilePath, err)
	}

	if fileInfo.IsDir() {
		return nil
	}

	if r.isImageFile(mediaPart.FilePath) {
		fileName := filepath.Base(mediaPart.FilePath)
		fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]

		item := database.ItemMetadata{
			Title:     fileName,
			LibraryID: library.ID,
			Library:   library,
			MediaPart: *mediaPart,
		}

		err = database.CreateImage(&item)
		if err != nil {
			return fmt.Errorf("could not resolve image metadata %s: %w", mediaPart.FilePath, err)
		}

		err = ants.Submit(func() {
			err := image.GetInformation(&item, library)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to get image information for %s: %w", mediaPart.FilePath, err)
			}
		})
		if err != nil {
			return fmt.Errorf("could not schedule image information job %s: %w", mediaPart.FilePath, err)
		}
	}

	return nil
}

func (r Resolver) isImageFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.StringInSlice(ext, supportedImageFormats)
}
