package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	"github.com/rs/zerolog/log"
)

// GetSupportedImageFormats returns a list of the image formats supported for ingestion.
// Non-supported common formats needing support from libvips are commented out.
// TODO: Check support for RAW formats.
func GetSupportedImageFormats() []string {
	return []string{
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
}

func IsImageFile(path string) bool {
	ext := filepath.Ext(path)

	return IsStringInSlice(ext, GetSupportedImageFormats())
}

func IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func IsFileReadable(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}

func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		log.Err(err).Msgf("Failed to get file size for %s", path)

		return 0, fmt.Errorf("failed to get file size for %w", err)
	}

	return info.Size(), nil
}

func HashFilePath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, writeFileHashErr := h.Write([]byte(path)); writeFileHashErr != nil {
		return nil, fmt.Errorf("could not write to hash: %w", writeFileHashErr)
	}

	return h.Sum(nil), nil
}

func HashFileBytes(fileBytes []byte) ([]byte, error) {
	h := sha256.New()
	if _, writeBytesHashErr := h.Write(fileBytes); writeBytesHashErr != nil {
		return nil, fmt.Errorf("could not write to hash: %w", writeBytesHashErr)
	}

	return h.Sum(nil), nil
}

func ChunkMediaSlice(slice []database.ItemMetadata, chunkSize int) [][]database.ItemMetadata {
	var chunks [][]database.ItemMetadata

	for {
		if len(slice) == 0 {
			break
		}

		// necessary check to avoid slicing beyond
		// slice capacity
		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}

	return chunks
}
