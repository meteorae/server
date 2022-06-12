package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// Supported image formats for ingestion. Non-supported common formats needing support from libvips are commented out.
// TODO: Check support for RAW formats.
var SupportedImageFormats = []string{
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

func IsImageFile(path string) bool {
	ext := filepath.Ext(path)

	return IsStringInSlice(ext, SupportedImageFormats)
}

func HashFilePath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := h.Write([]byte(path)); err != nil {
		return nil, fmt.Errorf("could not write to hash: %w", err)
	}

	return h.Sum(nil), nil
}

func HashFileBytes(fileBytes []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(fileBytes); err != nil {
		return nil, fmt.Errorf("could not write to hash: %w", err)
	}

	return h.Sum(nil), nil
}
