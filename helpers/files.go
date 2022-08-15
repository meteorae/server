package helpers

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
)

const (
	BaseDirectoryMode        = 0o755
	BaseFileMode             = 0o600
	BaseDirectoryPermissions = os.FileMode(BaseDirectoryMode)
	BaseFilePermissions      = os.FileMode(BaseFileMode)
	imageDownloadTimeout     = 10 * time.Second
)

func GetIgnoredFileGlobs() []string {
	return []string{
		// Unix hidden files, includes macOS-specific files
		"**/.*",

		// Sample files
		"**/sample.?",
		"**/sample.??",
		"**/sample.???",  // Matches sample.mkv
		"**/sample.????", // Matches sample.webm
		"**/sample.?????",
		"**/*.sample.?",
		"**/*.sample.??",
		"**/*.sample.???",
		"**/*.sample.????",
		"**/*.sample.?????",
		"**/sample/*",

		// Metadata directories
		"**/metadata/**",
		"**/metadata",

		// Kodi-compatible metadata
		"**/extrafanart/**",
		"**/extrafanart",
		"**/extrathumbs/**",
		"**/extrathumbs",
		"**/.actors/**",
		"**/.actors",

		// Western Digital directories
		"**/.wd_tv/**",
		"**/.wd_tv",

		// Unix lost files
		"**/lost+found/**",
		"**/lost+found",

		// Synology
		"**/eaDir/**",
		"**/eaDir",
		"**/@eaDir/**",
		"**/@eaDir",
		"**/#recycle/**",
		"**/#recycle",

		// Qnap
		"**/@Recycle/**",
		"**/@Recycle",
		"**/.@__thumb/**",
		"**/.@__thumb",

		// Windows
		"**/$RECYCLE.BIN/**",
		"**/$RECYCLE.BIN",
		"**/System Volume Information/**",
		"**/System Volume Information",

		// Windows thumbnail cache
		"**/thumbs.db",

		// Resilio directories
		"**/*.bts",
		"**/*.sync",
	}
}

// Given a path and a DirEntry, returns whether the given path should be ignored.
func ShouldIgnore(path string, d fs.DirEntry) bool {
	isMatched := false

	for _, ext := range GetIgnoredFileGlobs() {
		match, err := doublestar.Match(ext, path)
		if err != nil {
			// If the glob fails, be safe and don't ignore the file
			break
		}

		if match {
			isMatched = true

			break
		}
	}

	return isMatched
}

func EnsurePathExists(path string) error {
	return fmt.Errorf("failed to ensure path exists: %w", os.MkdirAll(path, BaseDirectoryPermissions))
}

// Saves a remote image file to the image cache.
// Returns the hash of the image file.
func SaveExternalImageToCache(fileURL, agent string, item sdk.Item, filetype string) (string, error) {
	var fileBuffer bytes.Buffer

	ctx, cancel := context.WithTimeout(context.TODO(), imageDownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download web client: %w", err)
	}
	defer res.Body.Close()

	// TODO: We should check if the file is an image before we save it.

	_, err = io.Copy(&fileBuffer, res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy remote image into memory: %w", err)
	}

	return SaveImageToCache(fileBuffer.Bytes(), agent, item, filetype)
}

// Internal method to generate the hash of the image file and save it to the cache.
// Returns the hash of the image file.
func SaveImageToCache(file []byte, agent string, item sdk.Item, filetype string) (string, error) {
	hash, err := utils.HashFileBytes(file)
	if err != nil {
		return "", fmt.Errorf("failed to hash remote image file: %w", err)
	}

	fileHash := hex.EncodeToString(hash)

	imageCachePath := metadata.GetFilepathForAgentAndHash(
		agent,
		fileHash,
		item.GetUUID().String(),
		item.GetType(),
		filetype)

	fileBuffer := bytes.NewBuffer(file)

	image, err := vips.NewImageFromReader(fileBuffer)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	export, _, err := image.ExportWebp(vips.NewWebpExportParams())
	if err != nil {
		return "", fmt.Errorf("failed to set image format: %w", err)
	}

	err = os.WriteFile(imageCachePath, export, BaseFilePermissions)
	if err != nil {
		return "", fmt.Errorf("failed to write image to disk: %w", err)
	}

	return fileHash, nil
}
