package scanner

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/scanners"
	"github.com/meteorae/meteorae-server/scanners/audio"
	"github.com/meteorae/meteorae-server/scanners/photos"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
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
	var (
		items     []database.ItemMetadata
		itemsInfo []sdk.Item
		mediaList []sdk.Item
	)

	fullPath := filepath.Join(root, directory)

	directoryContent, err := ioutil.ReadDir(fullPath)
	if err != nil {
		log.Err(err).Msgf("Failed to read directory %s", fullPath)

		return
	}

	files, dirs := filterFiles(directoryContent)

	var extensions []string

	switch library.Type {
	case database.MovieLibrary:
		extensions = video.VideoFileExtensions
	case database.TVLibrary:
		extensions = video.VideoFileExtensions
	case database.MusicLibrary:
		extensions = audio.AudioFileExtensions
	case database.ImageLibrary:
		extensions = append(photos.PhotoFileExtensions, video.VideoFileExtensions...)
	}

	scanFunc := scanners.GetScanFuncByName(library.Type.String(), library.Scanner)

	if scanFunc != nil {
		scanFunc(directory, &files, &dirs, &mediaList, extensions, root)
	} else {
		log.Err(errors.New("Scanner not found")).Msgf("Failed to scan directory %s", fullPath)

		return
	}

	// Check if files are already in the database. We don't want to add stuff twice.
	var existingMedia []database.ItemMetadata

	for _, media := range mediaList {
		if media, ok := media.(sdk.ItemInfo); ok {
			item, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(media.Parts[0])))
			if err != nil {
				log.Err(err).Msgf("Failed to get media by path %s", media.Parts[0])

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
			if m, ok := m.(sdk.ItemInfo); ok {
				if uuid.NewSHA1(library.UUID, []byte(m.Parts[0])) == media.UUID {
					mediaList = append(mediaList[:i], mediaList[i+1:]...)

					break
				}
			}
		}
	}

	// TODO: Figure out how to handle deleted items.

	if len(mediaList) > 0 {
		// Iterate over mediaList and assert types, then convert to database.ItemMetadata
		for _, media := range mediaList {
			if media, ok := media.(sdk.Movie); ok {
				// Check if item exists in database, to avoid scanning it in twice.
				if _, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath))); err != nil {
					// If item doesn't exist, add it.
					if errors.Is(err, gorm.ErrRecordNotFound) {
						parts := make([]database.MediaPart, 0, len(media.Parts))

						for _, part := range media.Parts {
							parts = append(parts, database.MediaPart{
								FilePath: part,
							})
						}

						items = append(items, database.ItemMetadata{
							Title:       media.Title,
							ReleaseDate: media.ReleaseDate,
							LibraryID:   library.ID,
							Parts:       parts,
							UUID:        uuid.NewSHA1(library.UUID, []byte(media.Parts[0])),
							Type:        sdk.MovieItem,
						})

						itemsInfo = append(itemsInfo, sdk.Movie{
							ItemInfo: &sdk.ItemInfo{
								Title:       media.Title,
								ReleaseDate: media.ReleaseDate,
								Parts:       media.Parts,
								UUID:        uuid.NewSHA1(library.UUID, []byte(media.Parts[0])),
							},
						})

						continue
					} else {
						log.Err(err).Msgf("Failed to get item by UUID %s", media.UUID)

						continue
					}
				} else {
					// If item exists, just skip it for now.
					log.Debug().Msgf("Skipping %s, already in database.", media.Parts[0])

					continue
				}
			}

			if media, ok := media.(sdk.TVEpisode); ok {
				// Check if item exists in database, to avoid scanning it in twice.
				if _, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath))); err != nil {
				} else {
					// If item exists, just skip it for now.
					// TODO: We should schedule a metadata refresh here.
					log.Debug().Msgf("Skipping %s, already in database.", media.Parts[0])

					continue
				}

				var (
					series database.ItemMetadata
					season database.ItemMetadata
				)

				// Check if the series exists in the database.
				// If not, create it.
				series, err := database.GetItemByTitleAndType(media.SeriesTitle, sdk.TVShowItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					series = database.ItemMetadata{
						Title:     media.SeriesTitle,
						LibraryID: library.ID,
						Type:      sdk.TVShowItem,
						UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					}

					series, err = database.CreateItem(series)
					if err != nil {
						log.Err(err).Msgf("Failed to create series %s", series.Title)

						continue
					}

					// Save the combined metadata.
					err := metadata.SaveMetadataToXML(sdk.TVShow{
						ItemInfo: &sdk.ItemInfo{
							ID:    series.ID,
							Title: series.Title,
							UUID:  series.UUID,
						},
					}, series.Type, "combined")
					if err != nil {
						log.Err(err).Msgf("Failed to save combined metadata for %s", series.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get series by title %s", media.Title)

					continue
				}

				// Check if the album exists in the database.
				// If not, create it.
				season, err = database.GetItemByParentWithIndex(series.ID, media.Season)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					season = database.ItemMetadata{
						Title:     fmt.Sprintf("Season %d", media.Season),
						ParentID:  series.ID,
						Sequence:  media.Season,
						LibraryID: library.ID,
						Type:      sdk.TVSeasonItem,
						UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					}

					season, err = database.CreateItem(season)
					if err != nil {
						log.Err(err).Msgf("Failed to create season %d for series %s", media.Season, series.Title)

						continue
					}

					// Save the combined metadata.
					err := metadata.SaveMetadataToXML(sdk.TVSeason{
						ItemInfo: &sdk.ItemInfo{
							ID:    series.ID,
							Title: series.Title,
							UUID:  series.UUID,
						},
					}, series.Type, "combined")
					if err != nil {
						log.Err(err).Msgf("Failed to save combined metadata for %s", series.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get season %d for series %s", media.Season, series.Title)

					continue
				}

				parts := make([]database.MediaPart, 0, len(media.Parts))

				for _, part := range media.Parts {
					parts = append(parts, database.MediaPart{
						FilePath: part,
					})
				}

				items = append(items, database.ItemMetadata{
					Title:     media.Title,
					ParentID:  season.ID,
					Sequence:  media.Episode,
					LibraryID: library.ID,
					Parts:     parts,
					UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					Type:      sdk.TVEpisodeItem,
				})

				itemsInfo = append(itemsInfo, sdk.TVEpisode{
					ItemInfo: &sdk.ItemInfo{
						Title: media.Title,
						Parts: media.Parts,
						UUID:  uuid.NewSHA1(library.UUID, []byte(fullPath)),
					},
					SeriesTitle: series.Title,
					SeriesID:    series.ID,
					SeasonTitle: season.Title,
					SeasonID:    season.ID,
					Season:      media.Season,
					Episode:     media.Episode,
				})
			}

			if media, ok := media.(sdk.MusicTrack); ok {
				var (
					artist database.ItemMetadata
					album  database.ItemMetadata
					medium database.ItemMetadata
				)

				// Check if the album artist exists in the database.
				// If not, create it.
				artist, err = database.GetItemByTitleAndType(media.AlbumArtist, sdk.PersonItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					artist = database.ItemMetadata{
						Title:     media.AlbumArtist,
						LibraryID: library.ID,
						Type:      sdk.PersonItem,
						UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					}

					var hasMusicBrainzArtistID bool

					for _, identifier := range artist.ExternalIdentifiers {
						if identifier.IdentifierType == sdk.MusicbrainzArtistIdentifier {
							hasMusicBrainzArtistID = true

							break
						}
					}

					if !hasMusicBrainzArtistID {
						var musicBrainzArtistID string

						for _, identifier := range media.Identifiers {
							if identifier.IdentifierType == sdk.MusicbrainzArtistIdentifier {
								musicBrainzArtistID = identifier.Identifier

								break
							}
						}

						artist.ExternalIdentifiers = append(artist.ExternalIdentifiers, database.ExternalIdentifier{
							Identifier:     strings.Trim(strings.TrimSpace(musicBrainzArtistID), "\x00"),
							IdentifierType: sdk.MusicbrainzArtistIdentifier,
						})
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
				album, err = database.GetItemByTitleAndType(media.AlbumName, sdk.MusicAlbumItem)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					album = database.ItemMetadata{
						Title:     media.AlbumName,
						ParentID:  artist.ID,
						LibraryID: library.ID,
						Type:      sdk.MusicAlbumItem,
						UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					}

					var hasMusicBrainzReleaseID bool

					for _, identifier := range album.ExternalIdentifiers {
						if identifier.IdentifierType == sdk.MusicbrainzReleaseIdentifier {
							hasMusicBrainzReleaseID = true

							break
						}
					}

					if !hasMusicBrainzReleaseID {
						var musicBrainzReleaseID string

						for _, identifier := range media.Identifiers {
							if identifier.IdentifierType == sdk.MusicbrainzArtistIdentifier {
								musicBrainzReleaseID = identifier.Identifier

								break
							}
						}

						album.ExternalIdentifiers = append(album.ExternalIdentifiers, database.ExternalIdentifier{
							Identifier:     strings.Trim(strings.TrimSpace(musicBrainzReleaseID), "\x00"),
							IdentifierType: sdk.MusicbrainzReleaseIdentifier,
						})
					}

					album, err = database.CreateItem(album)
					if err != nil {
						log.Err(err).Msgf("Failed to create album %s", album.Title)

						continue
					}

					// Save the combined metadata.
					// TODO: Add identifiers
					err := metadata.SaveMetadataToXML(sdk.MusicAlbum{
						ItemInfo: &sdk.ItemInfo{
							ID:    album.ID,
							Title: album.Title,
							UUID:  album.UUID,
						},
					}, album.Type, "combined")
					if err != nil {
						log.Err(err).Msgf("Failed to save combined metadata for %s", album.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get album by title %s", media.AlbumName)

					continue
				}

				// Check if the medium exists in the database.
				// If not, create it.
				medium, err = database.GetItemByParentWithIndex(album.ID, media.DiscIndex)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					medium = database.ItemMetadata{
						Title:     fmt.Sprintf("Disc %d", media.DiscIndex),
						ParentID:  album.ID,
						Sequence:  media.DiscIndex,
						LibraryID: library.ID,
						Type:      sdk.MusicMediumItem,
						UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					}

					medium, err = database.CreateItem(medium)
					if err != nil {
						log.Err(err).Msgf("Failed to create medium %s", medium.Title)

						continue
					}
				} else if err != nil {
					log.Err(err).Msgf("Failed to get medium for album %s", media.AlbumName)

					continue
				}

				parts := make([]database.MediaPart, 0, len(media.Parts))

				for _, part := range media.Parts {
					parts = append(parts, database.MediaPart{
						FilePath: part,
					})
				}

				items = append(items, database.ItemMetadata{
					Title:     media.Title,
					ParentID:  medium.ID,
					Sequence:  media.TrackIndex,
					LibraryID: library.ID,
					Parts:     parts,
					UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					Type:      sdk.MusicTrackItem,
				})
			}

			if media, ok := media.(sdk.Image); ok {
				paths := strings.Split(directory, string(os.PathSeparator))

				// If we're in the library root, we don't have a parent.
				var parentAlbumID uint
				if len(paths) == 1 && paths[0] == "." {
					parentAlbumID = 0
				} else if len(paths) == 2 {
					// Figure out if the parent directory has an album.
					album, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath)))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						album = database.ItemMetadata{
							Title:     paths[1],
							LibraryID: library.ID,
							Type:      sdk.ImageAlbumItem,
							UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
						}

						album, err = database.CreateItem(album)
						if err != nil {
							log.Err(err).Msgf("Failed to create album %s", album.Title)

							continue
						}
					} else if err != nil {
						log.Err(err).Msgf("Failed to get album for %s", fullPath)

						continue
					}

					parentAlbumID = album.ID
				} else {
					// Figure out if the parent directory has an album.
					album, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath)))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						parentDirectory := filepath.Dir(fullPath)

						// Figure out if the parent directory has an parentAlbum.
						parentAlbum, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(parentDirectory)))
						if err != nil {
							log.Err(err).Msgf("Failed to get album for path %s", parentDirectory)

							continue
						}

						album = database.ItemMetadata{
							Title:     paths[len(paths)-1],
							LibraryID: library.ID,
							ParentID:  parentAlbum.ID,
							Type:      sdk.ImageAlbumItem,
							UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
						}

						album, err = database.CreateItem(album)
						if err != nil {
							log.Err(err).Msgf("Failed to create album %s", album.Title)

							continue
						}
					} else if err != nil {
						log.Err(err).Msgf("Failed to get album for %s", fullPath)

						continue
					}

					parentAlbumID = album.ID
				}

				parts := make([]database.MediaPart, 0, len(media.Parts))

				for _, part := range media.Parts {
					parts = append(parts, database.MediaPart{
						FilePath: part,
					})
				}

				items = append(items, database.ItemMetadata{
					Title:     media.Title,
					ParentID:  parentAlbumID,
					LibraryID: library.ID,
					Parts:     parts,
					UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					Type:      sdk.ImageItem,
				})
			}

			if media, ok := media.(sdk.Video); ok {
				paths := strings.Split(directory, string(os.PathSeparator))

				// If we're in the library root, we don't have a parent.
				var parentAlbumID uint
				if len(paths) == 1 && paths[0] == "." {
					parentAlbumID = 0
				} else if len(paths) == 2 {
					// Figure out if the parent directory has an album.
					album, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath)))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						album = database.ItemMetadata{
							Title:     paths[1],
							LibraryID: library.ID,
							Type:      sdk.ImageAlbumItem,
							UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
						}

						album, err = database.CreateItem(album)
						if err != nil {
							log.Err(err).Msgf("Failed to create album %s", album.Title)

							continue
						}
					} else if err != nil {
						log.Err(err).Msgf("Failed to get album for %s", fullPath)

						continue
					}

					parentAlbumID = album.ID
				} else {
					// Figure out if the parent directory has an album.
					album, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(fullPath)))
					if errors.Is(err, gorm.ErrRecordNotFound) {
						parentDirectory := filepath.Dir(fullPath)

						// Figure out if the parent directory has an parentAlbum.
						parentAlbum, err := database.GetItemByUUID(uuid.NewSHA1(library.UUID, []byte(parentDirectory)))
						if err != nil {
							log.Err(err).Msgf("Failed to get album for path %s", parentDirectory)

							continue
						}

						album = database.ItemMetadata{
							Title:     paths[len(paths)-1],
							LibraryID: library.ID,
							ParentID:  parentAlbum.ID,
							Type:      sdk.ImageAlbumItem,
							UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
						}

						album, err = database.CreateItem(album)
						if err != nil {
							log.Err(err).Msgf("Failed to create album %s", album.Title)

							continue
						}
					} else if err != nil {
						log.Err(err).Msgf("Failed to get album for %s", fullPath)

						continue
					}

					parentAlbumID = album.ID
				}

				parts := make([]database.MediaPart, 0, len(media.Parts))

				for _, part := range media.Parts {
					parts = append(parts, database.MediaPart{
						FilePath: part,
					})
				}

				items = append(items, database.ItemMetadata{
					Title:     media.Title,
					ParentID:  parentAlbumID,
					LibraryID: library.ID,
					Parts:     parts,
					UUID:      uuid.NewSHA1(library.UUID, []byte(fullPath)),
					Type:      sdk.VideoClipItem,
				})
			}
		}

		chunkifiedItems := utils.ChunkMediaSlice(items, itemChunkSize)

		for _, chunk := range chunkifiedItems {
			err := database.CreateItemBatch(chunk)
			if err != nil {
				log.Err(err).Msg("Failed to create items")
			}
		}

		for i, item := range items {
			// TODO: We should probably check that these are the same?
			err := metadata.SaveMetadataToXML(itemsInfo[i], item.Type, "combined")
			if err != nil {
				log.Err(err).Msgf("Failed to save combined metadata for %s", item.Title)

				continue
			}
		}
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
