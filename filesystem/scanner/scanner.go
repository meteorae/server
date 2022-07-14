package scanner

import (
	"io/ioutil"
	"os"
	"path/filepath"

	movieagent "github.com/meteorae/meteorae-server/agents/movie"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/scanners/audio"
	movieScanner "github.com/meteorae/meteorae-server/scanners/movie"
	simplemovie "github.com/meteorae/meteorae-server/scanners/simpleMovie"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/panjf2000/ants/v2"
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

func scanDirectory(directory, root string, library database.Library) {
	var mediaList []database.ItemMetadata

	fullPath := filepath.Join(root, directory)

	directoryContent, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Err(err).Msgf("Failed to read directory %s", fullPath)

		return
	}

	files, dirs := filterFiles(directoryContent)

	switch library.Type {
	case database.MovieLibrary:
		movieScanner.Scan(directory, &files, &dirs, &mediaList, video.VideoFileExtensions, root)
	case database.MusicLibrary:
		audio.Scan(directory, &files, &dirs, &mediaList, audio.AudioFileExtensions, root)
	default:
		simplemovie.Scan(directory, &files, &dirs, &mediaList, video.VideoFileExtensions, root)
	}

	for _, dir := range dirs {
		scanDirectory(filepath.Join(directory, dir), root, library)
	}

	for i := range mediaList {
		mediaList[i].Library = library
	}

	// Check if files are already in the database. We don't want to add stuff twice.
	// Note: This has the side effect of preventing a directory from being in two libraries, as done currently.
	var existingMedia []database.ItemMetadata

	for _, media := range mediaList {
		item, err := database.GetMediaByPath(media.MediaParts[0].FilePath)
		if err != nil {
			log.Err(err).Msgf("Failed to get media by path %s", media.MediaParts[0].FilePath)

			continue
		}

		if item != nil {
			existingMedia = append(existingMedia, *item)
		}
	}

	for _, media := range existingMedia {
		for i, m := range mediaList {
			if m.ID == media.ID {
				mediaList = append(mediaList[:i], mediaList[i+1:]...)

				break
			}
		}
	}

	// TODO: Figure out how to handle deleted items.

	if len(mediaList) > 0 {
		err = database.CreateMovieBatch(&mediaList)
		if err != nil {
			log.Err(err).Msgf("Failed to create movie batch")
		}
	}

	// TODO: We may want to call a scheduled task instead, but we need scheduled tasks first.
	err = ants.Submit(func() {
		for _, movie := range mediaList {
			movieagent.Search(movie)
		}
	})
	if err != nil {
		log.Err(err).Msg("could not schedule agent job")
	}
}

func ScanDirectory(directory string, library database.Library) {
	if _, err := os.Lstat(directory); err != nil {
		log.Err(err).Msgf("Failed to scan directory %s", directory)

		return
	}

	scanDirectory(".", directory, library)
}
