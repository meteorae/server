package musicalbum

import (
	"errors"
	"fmt"

	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/michiwend/gomusicbrainz"
	"github.com/rs/zerolog/log"
	"go.uber.org/ratelimit"
)

const (
	requestLimit  = 15
	requestOffset = 0
)

var (
	errNoResultsFound      = errors.New("no results found")
	errUnsupportedItemType = errors.New("unsupported item type")
)

func GetSearchResults(item sdk.Item) ([]sdk.Item, error) {
	RateLimiter := ratelimit.New(1)

	MusicBrainzClient, musicBrainzClientErr := gomusicbrainz.NewWS2Client(
		"https://musicbrainz.org/ws/2",
		"Meteorae",
		helpers.Version,
		"https://meteorae.tv")

	if musicBrainzClientErr != nil {
		return nil, fmt.Errorf("Failed to create MusicBrainz client: %w", musicBrainzClientErr)
	}

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
		} else { //nolint:staticcheck // Temporarily commented out code.
			// searchQuery = fmt.Sprintf("%s AND artist:\"%s\"", searchQuery, artist.Title)
		}
	}

	log.Debug().Str("query", searchQuery).Uint("item_id", item.GetID()).Msgf("Searching album on MusicBrainz")

	RateLimiter.Take()

	searchResults, searchErr := MusicBrainzClient.SearchRelease(searchQuery, requestLimit, requestOffset)
	if searchErr != nil {
		return nil, fmt.Errorf("Failed to search MusicBrainz: %w", searchErr)
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
						Identifier:     string(album.ID),
					},
					{
						IdentifierType: sdk.MusicbrainzArtistIdentifier,
						Identifier:     artistID,
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
		return nil, errNoResultsFound
	}

	if album, ok := results[0].(sdk.MusicAlbum); ok {
		album.ID = item.GetID()
		album.UUID = item.GetUUID()

		return album, nil
	}

	return nil, errUnsupportedItemType
}
