package scanner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/meteorae/meteorae-server/database"
	simplemovie "github.com/meteorae/meteorae-server/scanners/simpleMovie"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

func filterFiles(files []os.FileInfo) ([]string, []string) {
	var filteredFiles []string

	var filteredDirs []string

	for _, file := range files {
		if file.IsDir() {
			filteredDirs = append(filteredDirs, file.Name())
		} else {
			filteredFiles = append(filteredFiles, file.Name())
		}
	}

	return filteredFiles, filteredDirs
}

func scanDirectory(directory, root string, mediaList *[]database.ItemMetadata) {
	fullPath := filepath.Join(root, directory)

	directoryContent, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Err(err).Msgf("Failed to read directory %s", fullPath)

		return
	}

	files, dirs := filterFiles(directoryContent)

	simplemovie.Scan(directory, &files, &dirs, mediaList, video.VideoFileExtensions, root)

	// TODO: Make this recursive.
	for _, dir := range dirs {
		scanDirectory(filepath.Join(directory, dir), root, mediaList)
	}

	log.Debug().Str("scanner", simplemovie.GetName()).Msgf("Found %d files in %s", len(files), directory)
	log.Debug().Str("scanner", simplemovie.GetName()).Msgf("Found %d directories in %s", len(dirs), directory)
}

func ScanDirectory(directory string, library database.Library) {
	defer utils.TimeTrack(time.Now())

	if _, err := os.Lstat(directory); err != nil {
		log.Err(err).Msgf("Failed to scan directory %s", directory)

		return
	}

	var mediaList []database.ItemMetadata

	scanDirectory(".", directory, &mediaList)
	log.Debug().Str("scanner", simplemovie.GetName()).Msgf("Found %d media items in %s", len(mediaList), directory)

	for i := range mediaList {
		mediaList[i].Library = library
	}

	err := database.CreateMovieBatch(&mediaList)
	if err != nil {
		log.Err(err).Msgf("Failed to create movie batch")
	}
}
