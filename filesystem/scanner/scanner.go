package scanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	movieagent "github.com/meteorae/meteorae-server/agents/movie"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/audio"
	movieScanner "github.com/meteorae/meteorae-server/scanners/movie"
	"github.com/meteorae/meteorae-server/scanners/music"
	"github.com/meteorae/meteorae-server/scanners/photos"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const itemChunkSize = 500

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
	var mediaList []model.Item

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
		music.Scan(directory, &files, &dirs, &mediaList, audio.AudioFileExtensions, root)
	case database.ImageLibrary:
		// Photo libraries also support video clips.
		imagesAndVideosExtensions := append(photos.PhotoFileExtensions, video.VideoFileExtensions...)

		photos.Scan(directory, &files, &dirs, &mediaList, imagesAndVideosExtensions, root)
	}

	// Check if files are already in the database. We don't want to add stuff twice.
	// Note: This has the side effect of preventing a directory from being in two libraries, as done currently.
	var existingMedia []database.ItemMetadata

	for _, media := range mediaList {
		if metadataMedia, ok := media.(models.MetadataModel); ok {
			item, err := database.GetItemByPath(metadataMedia.Parts[0].FilePath)
			if err != nil {
				log.Err(err).Msgf("Failed to get media by path %s", metadataMedia.Parts[0].FilePath)

				continue
			}

			// For easy comparison, if there is no title, we assume there is no item.
			if item.Title == "" {
				existingMedia = append(existingMedia, item)
			}
		}
	}

	for _, media := range existingMedia {
		for i, m := range mediaList {
			if m, ok := m.(models.MetadataModel); ok {
				if m.ID == media.ID {
					mediaList = append(mediaList[:i], mediaList[i+1:]...)

					break
				}
			}
		}
	}

	// TODO: Figure out how to handle deleted items.

	if len(mediaList) > 0 {
		// Iterate over mediaList and assert types, then convert to database.ItemMetadata
		var items []database.ItemMetadata

		for _, media := range mediaList {
			if media, ok := media.(models.Movie); ok {
				items = append(items, media.ToItemMetadata())
			}

			if media, ok := media.(models.Track); ok {
				var (
					artist database.ItemMetadata
					album  database.ItemMetadata
				)

				// Check if the album artist exists in the database.
				// If not, create it.
				artist, err = database.GetItemByTitleAndType(media.AlbumArtist, database.PersonItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					artist = database.ItemMetadata{
						Title: media.AlbumArtist,
						Type:  database.PersonItem,
					}

					artist, err = database.CreateItem(artist)
					if err != nil {
						log.Err(err).Msgf("Failed to create artist %s", artist.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get artist by title %s", media.AlbumArtist)

					continue
				}

				// Check if the album exists in the database.
				// If not, create it.
				album, err = database.GetItemByTitleAndType(media.AlbumName, database.MusicAlbumItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					album = database.ItemMetadata{
						Title:    media.AlbumName,
						ParentID: artist.ID,
						Type:     database.MusicAlbumItem,
					}

					album, err = database.CreateItem(album)
					if err != nil {
						log.Err(err).Msgf("Failed to create album %s", album.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get album by title %s", media.AlbumName)

					continue
				}

				media.AlbumID = album.ID

				// Check if the medium exists in the database.
				// If not, create it.
				medium, err := database.GetItemByParentWithIndex(album.ID, media.DiscIndex)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					medium = database.ItemMetadata{
						Title:    fmt.Sprintf("Disc %d", media.DiscIndex),
						ParentID: album.ID,
						Sequence: media.DiscIndex,
						Type:     database.MusicMediumItem,
					}

					medium, err = database.CreateItem(medium)
					if err != nil {
						log.Err(err).Msgf("Failed to create medium %s", medium.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get medium by ID %d", media.MediumID)

					continue
				}

				media.MediumID = medium.ID

				itemMetadata := media.ToItemMetadata()

				items = append(items, itemMetadata)
			}

			if media, ok := media.(models.Photo); ok {
				var album database.ItemMetadata

				// Just use the directory name for the album name.
				photoAlbumName := filepath.Base(directory)
				if photoAlbumName == "." {
					photoAlbumName = "Uncategorized"
				}

				// Check if a photo album exists in the database.
				// If not, create it.
				album, err := database.GetItemByTitleAndType(photoAlbumName, database.ImageAlbumItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					album = database.ItemMetadata{
						Title: photoAlbumName,
						Type:  database.ImageAlbumItem,
					}

					album, err = database.CreateItem(album)
					if err != nil {
						log.Err(err).Msgf("Failed to create album %s", album.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get album by title %s", photoAlbumName)

					continue
				}

				itemMetadata := media.ToItemMetadata()
				itemMetadata.ParentID = album.ID

				items = append(items, itemMetadata)
			}

			if media, ok := media.(models.VideoClip); ok {
				var album database.ItemMetadata

				// Just use the directory name for the album name.
				photoAlbumName := filepath.Base(directory)

				// Check if a photo album exists in the database.
				// If not, create it.
				album, err := database.GetItemByTitleAndType(photoAlbumName, database.ImageAlbumItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					album = database.ItemMetadata{
						Title: photoAlbumName,
						Type:  database.ImageAlbumItem,
					}

					album, err = database.CreateItem(album)
					if err != nil {
						log.Err(err).Msgf("Failed to create album %s", album.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get album by title %s", photoAlbumName)

					continue
				}

				itemMetadata := media.ToItemMetadata()
				itemMetadata.ParentID = album.ID

				items = append(items, itemMetadata)
			}
		}

		// If we have more than 500 items, break up the batch to avoid hitting SQLite's limits.
		if len(items) > itemChunkSize {
			chunkifiedItems := utils.ChunkMediaSlice(items, itemChunkSize)

			for _, chunk := range chunkifiedItems {
				err := database.CreateItemBatch(chunk)
				if err != nil {
					log.Err(err).Msg("Failed to create items")
				}
			}
		} else {
			err = database.CreateItemBatch(items)
			if err != nil {
				log.Err(err).Msgf("Failed to create movie batch")
			}
		}
	}

	// TODO: We may want to call a scheduled task instead, but we need scheduled tasks first.
	err = ants.Submit(func() {
		for _, movie := range mediaList {
			movieItem, ok := movie.(models.Movie)
			if ok {
				movieagent.Search(movieItem)
			}
		}
	})
	if err != nil {
		log.Err(err).Msg("could not schedule agent job")
	}

	for _, dir := range dirs {
		scanDirectory(filepath.Join(directory, dir), root, library)
	}
}

func ScanDirectory(directory string, library database.Library) {
	defer utils.TimeTrack(time.Now())

	if _, err := os.Lstat(directory); err != nil {
		log.Err(err).Msgf("Failed to scan directory %s", directory)

		return
	}

	scanDirectory(".", directory, library)
}
