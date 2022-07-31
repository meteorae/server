package musicalbum

import (
	"errors"
	"strings"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/michiwend/gomusicbrainz"
	"github.com/rs/zerolog/log"
	"github.com/shkh/lastfm-go/lastfm"
)

var (
	MusicBrainzClient *gomusicbrainz.WS2Client
	LastFMClient      *lastfm.Api
)

func init() {
	var err error
	MusicBrainzClient, err = gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"Meteorae",
		helpers.Version,
		"https://github.com/meteorae/server")

	if err != nil {
		panic(err)
	}

	LastFMClient = lastfm.New("f97dc0b009cf1a0788f8680fb3e69f0d", "b6ca4a9be2b03137117a108fd1bb68c4")
}

func GetName() string {
	return "Meteorae Music Agent"
}

func GetSearchResults(item database.ItemMetadata) ([]sdk.Item, error) {
	var results []sdk.Item

	searchResults, err := MusicBrainzClient.SearchRelease(item.Title, 15, 0)
	if err != nil {
		return nil, err
	}

	for _, album := range searchResults.Releases {
		results = append(results, sdk.MusicAlbum{
			ItemInfo: &sdk.ItemInfo{
				Title:       album.Title,
				ReleaseDate: album.Date.Time,
				MatchScore:  uint(searchResults.Scores[album]),
			},
			MusicBrainzAlbumID:  string(album.ID),
			MusicBrainzArtistID: string(album.ArtistCredit.NameCredits[0].Artist.ID),
		})
	}

	return results, nil
}

func GetMetadata(item database.ItemMetadata) (database.ItemMetadata, error) {
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

	// Check Last.fm for album info. We always have a MBID, so we can use that.
	lastFMResult, err := LastFMClient.Album.GetInfo(lastfm.P{
		"mbid": album.MusicBrainzAlbumID,
	})

	// The last image is usually the largest, so we'll use that.
	var (
		imageURL  string
		imageHash string
	)

	if len(lastFMResult.Images) > 0 {
		imageURL = lastFMResult.Images[len(lastFMResult.Images)-1].Url
	}

	// Last.fm sends us a 300x300 image, but we want a 1200x1200 one.
	if imageURL != "" {
		imageURL = strings.Replace(imageURL, "300x300", "1200x1200", 1)

		// Cache the image locally for future use.
		imageHash, err = helpers.SaveExternalImageToCache(imageURL)
		if err != nil {
			log.Err(err).Msgf("Failed to save image %s", imageURL)
		}
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
