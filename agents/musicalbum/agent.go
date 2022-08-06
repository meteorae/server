package musicalbum

import (
	"errors"
	"fmt"
	"strings"

	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/michiwend/gomusicbrainz"
	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
	"go.uber.org/ratelimit"
)

var (
	MusicBrainzClient *gomusicbrainz.WS2Client
	LastFMClient      *lastfm.Api
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

	RateLimiter = ratelimit.New(1) // Limit to 1 request per second, to try to get around the rate limit.
}

func GetIdentifier() string {
	return "tv.meteorae.agents.musicbrainz"
}

func GetName() string {
	return "Meteorae Music Agent"
}

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	var searchQuery string

	// Get the artist name.
	/*artist, err := database.GetItemByID(item.ParentID)
	if err != nil {
		return nil, err
	}*/

	// Do we have an artist ID?
	var artistID string

	/*for i := range artist.ExternalIdentifiers {
		if artist.ExternalIdentifiers[i].IdentifierType == sdk.MusicbrainzArtistIdentifier {
			artistID = artist.ExternalIdentifiers[i].Identifier
		}
	}*/

	// If we have a MusicBrainz ID, we can use that.
	for i := range item.GetIdentifiers() {
		if item.GetIdentifiers()[i].IdentifierType == sdk.MusicbrainzReleaseIdentifier {
			searchQuery = fmt.Sprintf("reid:%s", item.GetIdentifiers()[i].Identifier)
		}
	}

	// Otherwise, build a normal query.
	if searchQuery == "" {
		searchQuery = fmt.Sprintf("release:\"%s\"", item.GetTitle())

		if artistID != "" {
			searchQuery = fmt.Sprintf("%s AND arid:%s", searchQuery, artistID)
		} else {
			// searchQuery = fmt.Sprintf("%s AND artist:\"%s\"", searchQuery, artist.Title)
		}
	}

	log.Debug().Str("query", searchQuery).Uint("item_id", item.GetID()).Msgf("Searching album on MusicBrainz")

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
				Identifiers: []sdk.Identifier{
					{
						IdentifierType: sdk.MusicbrainzReleaseIdentifier,
						Identifier:     fmt.Sprintf("%s", album.ID),
					},
					{
						IdentifierType: sdk.MusicbrainzArtistIdentifier,
						Identifier:     fmt.Sprintf("%s", artistID),
					},
					{
						IdentifierType: sdk.MusicbrainzReleaseGroupIdentifier,
						Identifier:     string(album.ReleaseGroup.ID),
					},
				},
			},
			// TODO: Support localized aliases.
			AlbumArtist: album.ArtistCredit.NameCredits[0].Artist.Name,
			// ArtistThumb: artist.Thumb,
		})
	}

	return results, nil
}

func GetMetadata(item sdk.Item) (sdk.Item, error) {
	log.Debug().Uint("item_id", item.GetID()).Str("title", item.GetTitle()).Msgf("Getting metadata for album")

	results, err := GetSearchResults(item)
	if err != nil {
		log.Err(err).Msgf("Failed to get music album results %s", item.GetTitle())

		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("No results found")
	}

	if album, ok := results[0].(sdk.MusicAlbum); ok {
		// The last image is usually the largest, so we'll use that.
		var (
			imageURL  string
			imageHash string
		)

		imageURL = getLastFmURL(album)

		if imageURL != "" {
			// Cache the image locally for future use.
			imageHash, err = helpers.SaveExternalImageToCache(imageURL, GetIdentifier(), item, "thumb")
			if err != nil {
				log.Err(err).Msgf("Failed to save image %s", imageURL)
			}
		}

		// TODO: Move this to another agent.
		/*
			// If we didn't find anything, see if the tracks have images embedded.
			if imageHash == "" {
				medium, err := database.GetChildFromItem(item.ID, sdk.MusicMediumItem)
				if err != nil {
					log.Err(err).Msgf("Failed to get medium %s", item.Title)
				}

				if medium.Title != "" {
					track, err := database.GetChildFromItem(medium.ID, sdk.MusicTrackItem)
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
			}*/

		album.Thumb = sdk.Posters{
			Items: []sdk.ItemImage{
				{
					External:  true,
					Provider:  GetIdentifier(),
					Media:     metadata.GetURIForAgent(GetIdentifier(), imageHash),
					URL:       imageURL,
					SortOrder: 0,
				},
			},
		}

		album.UUID = item.GetUUID()

		return album, nil
	}

	return nil, errors.New("got unexpected item type")
}

func getLastFmURL(album sdk.MusicAlbum) string {
	var releaseID string

	// Get the Release Group and Release ID.
	for _, identifier := range album.ItemInfo.Identifiers {
		if identifier.IdentifierType == sdk.MusicbrainzReleaseIdentifier {
			releaseID = identifier.Identifier

			break
		}
	}

	var (
		lastFMResult lastfm.AlbumGetInfo
		err          error
	)

	log.Debug().
		Str("album", album.Title).
		Str("artist", album.AlbumArtist).
		Str("mbid", releaseID).
		Msgf("Getting last.fm album image")

	lastFMResult, err = LastFMClient.Album.GetInfo(lastfm.P{
		"artist": album.AlbumArtist,
		"album":  album.Title,
		"mbid":   releaseID,
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
