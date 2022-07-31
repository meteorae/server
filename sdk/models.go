// Package sdk provides common structs and plugin facilities for Meteorae.
package sdk

import (
	"time"

	"github.com/google/uuid"
)

type Item interface {
	IsItem()
}

type ItemInfo struct {
	UUID          uuid.UUID
	Title         string
	OriginalTitle string
	ReleaseDate   time.Time
	Parts         []string
	Thumb         string
	ArtUrl        string
	MatchScore    uint
	LibraryID     uint
}

func (i ItemInfo) IsItem() {}

type IdentifierType int8

const (
	ImdbIdentifier IdentifierType = iota
	TmdbIdentifier
	AnidbIdentifier
	TvdbIdentifier
	MusicbrainzArtistIdentifier
	MusicbrainzReleaseIdentifier
	FacebookIdentifier
	TwitterIdentifier
	InstagramIdentifier
)

func (d IdentifierType) String() string {
	return [...]string{
		"IMDB ID",
		"TheMovieDB ID",
		"AniDB ID",
		"TVDB ID",
		"MusicBrainz Artist ID",
		"MusicBrainz Release ID",
		"Facebook ID",
		"Twitter ID",
		"Instagram ID",
	}[d]
}

// A Movie represents information about an individual movie, obtained through scanning or through an agent.
type Movie struct {
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
	TmdbID int
}

// A TVShow represents information about a TV series, obtained through scanning or through an agent.
type TVShow struct {
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
	SeriesTitle string
	Season      int
	Episode     int
	TmdbID      int
	TvdbID      int
}

// A TVEpisode represents information about an individual episode of a TV series,
// obtained through scanning or through an agent.
type TVEpisode struct {
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
	SeriesTitle string
	Season      int
	Episode     int
	TmdbID      int
	TvdbID      int
}

type MusicAlbum struct {
	*ItemInfo
	AlbumArtist         string
	Artist              []string
	MusicBrainzAlbumID  string
	MusicBrainzArtistID string
}

// A MusicTrack represents information about an individual track of a music album,
// obtained through scanning or through an agent.
type MusicTrack struct {
	*ItemInfo
	AlbumArtist string
	AlbumName   string
	DiscIndex   int
	TrackIndex  int
	Artist      []string
}

type Image struct {
	*ItemInfo
}

type Video struct {
	*ItemInfo
}
