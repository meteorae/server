package musicalbum

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dhowden/tag"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/fanart"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/michiwend/gomusicbrainz"
	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
	"go.uber.org/ratelimit"
)

const FANART_TV_API_KEY = "84d310b84b0b62da0cb23f8355271442"

const FANART_ALBUM_URL = "https://webservice.fanart.tv/v3/music/albums/%s?api_key=%d"

var (
	MusicBrainzClient *gomusicbrainz.WS2Client
	LastFMClient      *lastfm.Api
	FanartClient      *fanart.Client
	RateLimiter       ratelimit.Limiter
)

func init() {
	var err error
	MusicBrainzClient, err = gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"Meteorae",
		helpers.Version,
		"https://meteorae.tv")

	if err != nil {
		panic(err)
	}

	LastFMClient = lastfm.New("f97dc0b009cf1a0788f8680fb3e69f0d", "b6ca4a9be2b03137117a108fd1bb68c4")

	FanartClient = fanart.New()

	RateLimiter = ratelimit.New(1) // Limit to 1 request per second, to try to get around the rate limit.
}

func GetName() string {
	return "Meteorae Music Agent"
}

func GetSearchResults(item database.ItemMetadata) ([]sdk.Item, error) {
	var searchQuery string

	// Get the artist name.
	artist, err := database.GetItemByID(item.ParentID)
	if err != nil {
		return nil, err
	}

	// If we have a MusicBrainz ID, we can use that.
	for i := range item.ExternalIdentifiers {
		if item.ExternalIdentifiers[i].IdentifierType == sdk.MusicbrainzReleaseIdentifier {
			searchQuery = fmt.Sprintf("reid:%s", item.ExternalIdentifiers[i].Identifier)
		}
	}

	// Otherwise, build a normal query.
	if searchQuery == "" {
		searchQuery = fmt.Sprintf("release:\"%s\"", item.Title)

		// Do we have an artist ID?
		var artistID string

		for i := range artist.ExternalIdentifiers {
			if artist.ExternalIdentifiers[i].IdentifierType == sdk.MusicbrainzArtistIdentifier {
				artistID = artist.ExternalIdentifiers[i].Identifier
			}
		}

		if artistID != "" {
			searchQuery = fmt.Sprintf("%s AND arid:%s", searchQuery, artistID)
		} else {
			searchQuery = fmt.Sprintf("%s AND artist:\"%s\"", searchQuery, artist.Title)
		}
	}

	log.Debug().Str("query", searchQuery).Uint("item_id", item.ID).Msgf("Searching album on MusicBrainz")

	RateLimiter.Take()

	searchResults, err := MusicBrainzClient.SearchRelease(searchQuery, 15, 0)
	if err != nil {
		return nil, err
	}

	results := make([]sdk.Item, 0, len(searchResults.Releases))

	for _, album := range searchResults.Releases {
		results = append(results, sdk.MusicAlbum{
			ItemInfo: &sdk.ItemInfo{
				Title:       album.Title,
				ReleaseDate: album.Date.Time,
				MatchScore:  uint(searchResults.Scores[album]),
			},
			// TODO: Support localized aliases.
			AlbumArtist:             album.ArtistCredit.NameCredits[0].Artist.Name,
			ArtistThumb:             artist.Thumb,
			MusicBrainzAlbumGroupID: string(album.ReleaseGroup.ID),
			MusicBrainzAlbumID:      string(album.ID),
			MusicBrainzArtistID:     string(album.ArtistCredit.NameCredits[0].Artist.ID),
		})
	}

	return results, nil
}

func GetMetadata(item database.ItemMetadata) (database.ItemMetadata, error) {
	log.Debug().Uint("item_id", item.ID).Str("title", item.Title).Msgf("Getting metadata for album")

	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to get music album results %s", item.Title)

		return database.ItemMetadata{}, err
	}

	if len(results) == 0 {
		return database.ItemMetadata{}, nil
	}

	album, ok := results[0].(sdk.MusicAlbum)
	if !ok {
		return database.ItemMetadata{}, errors.New("invalid result type")
	}

	// The last image is usually the largest, so we'll use that.
	var (
		imageURL  string
		imageHash string
	)

	// Try Fanart.TV first for the image.
	imageURL = getFanartUrl(album)

	// If we didn't find anything, fallback to Last.fm
	if imageURL == "" {
		imageURL = getLastFmURL(album)
	}

	if imageURL != "" {
		// Cache the image locally for future use.
		imageHash, err = helpers.SaveExternalImageToCache(imageURL)
		if err != nil {
			log.Err(err).Msgf("Failed to save image %s", imageURL)
		}
	}

	// If we didn't find anything, see if the tracks have images embedded.
	if imageHash == "" {
		medium, err := database.GetChildFromItem(item.ID, database.MusicMediumItem)
		if err != nil {
			log.Err(err).Msgf("Failed to get medium %s", item.Title)
		}

		if medium.Title != "" {
			track, err := database.GetChildFromItem(medium.ID, database.MusicTrackItem)
			if err != nil {
				log.Err(err).Msgf("Failed to get track %s", item.Title)
			}

			mediaParts, err := database.GetMediaParts(track.ID)
			if err != nil {
				log.Err(err).Msgf("Failed to get media parts %s", item.Title)
			}

			if len(mediaParts) > 0 {
				mediaFile, err := os.Open(mediaParts[0].FilePath)
				if err != nil {
					log.Err(err).Msgf("Failed to open file %s", track.Parts[0].FilePath)
				}
				defer mediaFile.Close()

				if mediaFile != nil {
					metadata, err := tag.ReadFrom(mediaFile)
					if err != nil {
						log.Err(err).Msgf("Failed to read tags from file %s", track.Parts[0].FilePath)
					}

					if metadata != nil {
						imageData := metadata.Picture().Data

						if imageData != nil {
							imageHash, err = helpers.SaveImageToCache(imageData)
							if err != nil {
								log.Err(err).Msgf("Failed to save image %s", track.Parts[0].FilePath)
							}
						}
					}
				}
			}
		}
	}

	// If we *still* didn't find anything, fallback to the parent artist's image.
	if imageHash == "" {
		imageHash = album.ArtistThumb
	}

	return database.ItemMetadata{
		Title:       album.ItemInfo.Title,
		ReleaseDate: album.ItemInfo.ReleaseDate,
		Thumb:       imageHash,
		ExternalIdentifiers: []database.ExternalIdentifier{
			{
				IdentifierType: sdk.MusicbrainzReleaseIdentifier,
				Identifier:     album.MusicBrainzAlbumID,
			},
			{
				IdentifierType: sdk.MusicbrainzArtistIdentifier,
				Identifier:     album.MusicBrainzArtistID,
			},
		},
	}, nil
}

type FanartAlbum struct{}

func getFanartUrl(album sdk.MusicAlbum) string {
	log.Debug().
		Str("album", album.Title).
		Str("mbid", album.MusicBrainzAlbumGroupID).
		Msgf("Getting fanart.tv album image")

	albumGroupInfo, err := FanartClient.GetAlbumImages(album.MusicBrainzAlbumGroupID)
	if err != nil {
		log.Err(err).Msgf("Failed to get album info from Fanart.tv for %s", album.ItemInfo.Title)

		return ""
	}

	if len(albumGroupInfo.Albums) == 0 {
		log.Debug().Msgf("No album info found on Fanart.tv for %s", album.ItemInfo.Title)

		return ""
	}

	if albumInfo, ok := albumGroupInfo.Albums[album.MusicBrainzAlbumID]; ok {
		if len(albumInfo.AlbumCover) > 0 {
			return fanart.GetBestImage(albumInfo.AlbumCover).URL
		}

		log.Debug().Msgf("No album cover found on Fanart.tv for %s", album.ItemInfo.Title)
	}

	return ""
}

func getLastFmURL(album sdk.MusicAlbum) string {
	var (
		lastFMResult lastfm.AlbumGetInfo
		err          error
	)

	log.Debug().
		Str("album", album.Title).
		Str("artist", album.AlbumArtist).
		Str("mbid", album.MusicBrainzAlbumID).
		Msgf("Getting last.fm album image")

	lastFMResult, err = LastFMClient.Album.GetInfo(lastfm.P{
		"artist": album.AlbumArtist,
		"album":  album.Title,
		"mbid":   album.MusicBrainzAlbumID,
	})
	if err != nil {
		log.Err(err).Msgf("Failed to get album info from Last.fm for %s", album.Title)

		return ""
	}

	if len(lastFMResult.Images) > 0 {
		// Last.fm sends us a 300x300 image, but we want a 1200x1200 one.
		imageURL := strings.Replace(lastFMResult.Images[len(lastFMResult.Images)-1].Url, "300x300", "1200x1200", 1)

		return imageURL
	}

	return ""
}
