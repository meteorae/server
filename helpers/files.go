package helpers

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/meteorae/meteorae-server/utils"
)

var (
	BaseDirectoryMode = 0o755
	BaseFileMode      = 0o600
)

var (
	BaseDirectoryPermissions = os.FileMode(BaseDirectoryMode)
	BaseFilePermissions      = os.FileMode(BaseFileMode)
)

var AudioFileExtensions = []string{
	".nsv",
	".m4a",
	".flac",
	".aac",
	".strm",
	".pls",
	".rm",
	".mpa",
	".wav",
	".wma",
	".ogg",
	".opus",
	".mp3",
	".mp2",
	".mod",
	".amf",
	".669",
	".dmf",
	".dsm",
	".far",
	".gdm",
	".imf",
	".it",
	".m15",
	".med",
	".okt",
	".s3m",
	".stm",
	".sfx",
	".ult",
	".uni",
	".xm",
	".sid",
	".ac3",
	".dts",
	".cue",
	".aif",
	".aiff",
	".ape",
	".mac",
	".mpc",
	".mp+",
	".mpp",
	".shn",
	".wv",
	".nsf",
	".spc",
	".gym",
	".adplug",
	".adx",
	".dsp",
	".adp",
	".ymf",
	".ast",
	".afc",
	".hps",
	".xsp",
	".acc",
	".m4b",
	".oga",
	".dsf",
	".mka",
}

var BookFileExtensions = []string{
	".azw",
	".azw3",
	".cb7",
	".cbr",
	".cbt",
	".cbz",
	".epub",
	".mobi",
	".pdf",
}

var IgnoredFileGlobs = []string{
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

// Given a path and a DirEntry, returns whether the given path should be ignored.
func ShouldIgnore(path string, d fs.DirEntry) bool {
	isMatched := false

	for _, ext := range IgnoredFileGlobs {
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

// Saves a local image file to the image cache.
// Returns the hash of the image file.
func SaveLocalImageToCache(filePath string) (string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open local image file: %w", err)
	}

	return saveImageToCache(file)
}

// Saves a remote image file to the image cache.
// Returns the hash of the image file.
func SaveExternalImageToCache(filePath string) (string, error) {
	var fileBuffer bytes.Buffer

	response, err := http.Get(filePath) //#nosec
	if err != nil {
		return "", fmt.Errorf("failed to fetch image \"%s\": %w", filePath, err)
	}
	defer response.Body.Close()

	_, err = io.Copy(&fileBuffer, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy remote image into memory: %w", err)
	}

	return saveImageToCache(fileBuffer.Bytes())
}

// Internal method to generate the hash of the image file and save it to the cache.
// Returns the hash of the image file.
func saveImageToCache(file []byte) (string, error) {
	hash, err := utils.HashFileBytes(file)
	if err != nil {
		return "", fmt.Errorf("failed to hash remote image file: %w", err)
	}

	fileHash := hex.EncodeToString(hash)
	prefix := fileHash[0:2]

	imageCachePath, err := xdg.CacheFile("meteorae/images")
	if err != nil {
		return "", fmt.Errorf("failed to get image cache path: %w", err)
	}

	fileBuffer := bytes.NewBuffer(file)

	image, err := vips.NewImageFromReader(fileBuffer)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	export, _, err := image.ExportWebp(vips.NewWebpExportParams())
	if err != nil {
		return "", fmt.Errorf("failed to set image format: %w", err)
	}

	cachedFilePath := filepath.Join(imageCachePath, prefix, fileHash)

	err = os.MkdirAll(cachedFilePath, BaseDirectoryPermissions)
	if err != nil {
		return "", fmt.Errorf("failed to create image cache directory: %w", err)
	}

	cachedFilePath = filepath.Join(cachedFilePath, "0x0.webp")

	err = ioutil.WriteFile(cachedFilePath, export, BaseFilePermissions)
	if err != nil {
		return "", fmt.Errorf("failed to write image to disk: %w", err)
	}

	return fileHash, nil
}
