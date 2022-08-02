package filter

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

var (
	IgnoredDirs = []string{
		`@eaDir`,
		`.*_UNPACK_.*`,
		`.*_FAILED_.*`,
		`\..*`,
		`lost\+found`,
		`.AppleDouble`,
		`.*\.itlp$`,
		`@Recycle`,
		`.*\.photoslibrary`,
		`#recycle`,
		`@Recently-Snapshot`,
	}
	RootIgnoredDirs = []string{
		`\$Recycle.Bin`,
		`System Volume Information`,
		`Temporary Items`,
		`Network Trash Folder`,
	}
)

func GetName() string {
	return "Filter Scanner"
}

// ParsePlexIgnore parses the .plexignore file at the given file path for ignored files and directories.
// This is a PLEX COMPATIBILITY feature.
func ParsePlexIgnore(filename string, ignoredFiles, ignoredDirs *[]string, cwd string) {
	file, err := os.Open(filepath.Join(filename, ".plexignore"))
	if err != nil {
		log.Err(err).Msg("Could not open .plexignore file")

		return
	}
	defer file.Close()

	// Parse the .plexignore file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := scanner.Text()
		pattern = strings.Trim(pattern, " ")

		// Ignore empty lines and comments
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			if strings.HasPrefix(pattern, "\\#") {
				pattern = strings.TrimPrefix(pattern, "\\")
			}

			if !strings.Contains(pattern, "/") {
				// We'll add this pattern as-is for now and use DoubleStar later to resolve it.
				*ignoredFiles = append(*ignoredFiles, pattern)
			} else {
				// Paths are relative to the directory of the .plexignore file.
				// To ensure resolving works properly, remove the leading slash.
				pattern = strings.TrimPrefix(pattern, "/")

				rootDir := filepath.Dir(filename)
				if cwd != "" {
					rootDir = cwd
				}

				*ignoredDirs = append(*ignoredDirs, filepath.Join(rootDir, pattern))
			}
		}
	}
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	filesToRemove := []string{}

	plexIgnoreFiles := []string{}
	plexIgnoreDirs := []string{}

	// PLEX COMPATIBILITY: Check for ignored files using .plexignore
	if root != "" && utils.IsStringInSlice(".plexignore", *files) {
		ParsePlexIgnore(filepath.Join(root, path), &plexIgnoreFiles, &plexIgnoreDirs, "")
	}

	// PLEX COMPATIBILITY: Also check in the root folder for .plexignore
	if root != "" && (!(len(*files) == 0) || (len(*files) > 0 && root != filepath.Dir((*files)[0]))) {
		if utils.IsFileExists(filepath.Join(root, ".plexignore")) {
			ParsePlexIgnore(root, &plexIgnoreFiles, &plexIgnoreDirs, filepath.Join(root, path))
		}
	}

	for _, file := range *files {
		fullPath := filepath.Join(root, path, file)

		// Get filename without extension and extension
		filename := filepath.Base(file)
		extension := strings.TrimLeft(filepath.Ext(filename), ".")
		filename = strings.TrimSuffix(filename, extension)

		// Check if the extension is supported
		if !utils.IsStringInSlice(extension, extensions) {
			filesToRemove = append(filesToRemove, file)
		}

		// Filter out broken symlinks and empty files
		fileSize, _ := utils.GetFileSize(fullPath)
		if !utils.IsFileExists(fullPath) || fileSize == 0 {
			filesToRemove = append(filesToRemove, file)
		}

		// Remove unreadable files
		if !utils.IsFileReadable(fullPath) {
			filesToRemove = append(filesToRemove, file)
		}

		// Remove hidden files
		if strings.HasPrefix(filename, ".") {
			filesToRemove = append(filesToRemove, file)
		}

		// Remove .plexignore ignored files
		for _, ignoredFile := range plexIgnoreFiles {
			match, err := doublestar.Match(ignoredFile, filepath.Base(path))
			if err != nil {
				log.Err(err).Msg("Could not match file against .plexignore")

				continue
			}

			if match {
				filesToRemove = append(filesToRemove, file)
			}
		}

		ignoredDirectories := IgnoredDirs
		if path == "" {
			ignoredDirectories = RootIgnoredDirs
		}

		dirsToRemove := []string{}

		for _, directory := range *dirs {
			baseDir := filepath.Base(directory)

			for _, ignoredDir := range ignoredDirectories {
				ignoredRegex := regexp.MustCompile(ignoredDir)
				if ignoredRegex.MatchString(baseDir) {
					dirsToRemove = append(dirsToRemove, directory)
				}
			}
		}

		// iterate over plexIgnoredDirs
		for _, ignoredDir := range plexIgnoreDirs {
			match, err := doublestar.Match(ignoredDir, filepath.Base(path))
			if err != nil {
				log.Err(err).Msg("Could not match directory against .plexignore")
			}

			if match {
				dirsToRemove = append(dirsToRemove, ignoredDir)
			}
		}

		for _, directoryToRemove := range dirsToRemove {
			*dirs = utils.RemoveStringFromSlice(directoryToRemove, *dirs)
		}

		for _, fileToRemove := range filesToRemove {
			*files = utils.RemoveStringFromSlice(fileToRemove, *files)
		}
	}
}
