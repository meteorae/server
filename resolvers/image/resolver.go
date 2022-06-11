package image

import (
	"fmt"
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

func (r Resolver) SupportsFileType(filePath string, isDir bool) bool {
	if isDir {
		return false
	}

	return r.isImageFile(filePath)
}

func (r Resolver) Resolve(mediaPart *database.MediaPart, library database.Library) error {
	fileName := filepath.Base(mediaPart.Path)
	fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]

	album, err := database.GetItemByPath(filepath.Dir(mediaPart.Path))
	if err != nil {
		return fmt.Errorf("failed to get image album for path %s: %w", mediaPart.Path, err)
	}

	item := &database.ItemMetadata{
		Title:     fileName,
		Type:      database.ImageItem,
		LibraryID: library.Id,
		Library:   library,
		ParentId:  album.Id,
		Path:      mediaPart.Path,
	}

	item, err = database.CreateItem(item)
	if err != nil {
		return fmt.Errorf("could not resolve image metadata %s: %w", mediaPart.Path, err)
	}

	mediaPart.ItemId = item.Id

	_, err = database.CreateMediaPart(*mediaPart)
	if err != nil {
		return fmt.Errorf("failed to create media part for %s: %w", mediaPart.Path, err)
	}

	err = ants.Submit(func() {
		err := image.GetInformation(item, library)
		if err != nil {
			log.Error().Err(err).Msgf("failed to get image information for %s: %w", mediaPart.Path, err)
		}
	})
	if err != nil {
		return fmt.Errorf("could not schedule image information job %s: %w", mediaPart.Path, err)
	}

	return nil
}

func (r Resolver) isImageFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.StringInSlice(ext, supportedImageFormats)
}
