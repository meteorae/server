// Package sdk provides common structs and plugin facilities for Meteorae.
package sdk

import (
	"encoding/xml"
	"time"

	"github.com/google/uuid"
)

type Item interface {
	IsItem()
	GetUUID() string
	GetIdentifiers() []Identifier
	GetThumbs() []ItemImage
	GetArt() []ItemImage
}

type ItemImage struct {
	XMLName   xml.Name `xml:"item"`
	External  bool     `xml:"external,attr"`
	URL       string   `xml:"url,attr"`
	Provider  string   `xml:"provider,attr"`
	Media     string   `xml:"media,attr"`
	SortOrder uint     `xml:"sort_order,attr"`
}

type Art struct {
	XMLName xml.Name    `xml:"art"`
	Items   []ItemImage `xml:"item"`
}

type Posters struct {
	XMLName xml.Name    `xml:"posters"`
	Items   []ItemImage `xml:"item"`
}

type ItemInfo struct {
	UUID          uuid.UUID
	Title         string       `xml:"title"`
	OriginalTitle string       `xml:"original_title"`
	ReleaseDate   time.Time    `xml:"release_date"`
	Parts         []string     `xml:"-"`
	Thumb         Posters      `xml:"posters"`
	Art           Art          `xml:"art"`
	Language      string       `xml:"language"`
	Identifiers   []Identifier `xml:"identifier"`
	MatchScore    uint         `xml:"-"`
	LibraryID     uint         `xml:"-"`
}

func (i ItemInfo) IsItem() {}

func (i ItemInfo) GetUUID() string {
	return i.UUID.String()
}

func (i ItemInfo) GetIdentifiers() []Identifier {
	return i.Identifiers
}

func (i ItemInfo) GetThumbs() []ItemImage {
	return i.Thumb.Items
}

func (i ItemInfo) GetArt() []ItemImage {
	return i.Art.Items
}

type IdentifierType int8

const (
	ImdbIdentifier IdentifierType = iota
	TmdbIdentifier
	AnidbIdentifier
	TvdbIdentifier
	MusicbrainzArtistIdentifier
	MusicbrainzReleaseIdentifier
	MusicbrainzReleaseGroupIdentifier
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
		"MusicBrainz Release Group ID",
		"Facebook ID",
		"Twitter ID",
		"Instagram ID",
	}[d]
}

type Identifier struct {
	XMLName        xml.Name       `xml:"external_identifier"`
	IdentifierType IdentifierType `xml:"type,attr"`
	Identifier     string         `xml:"value,attr"`
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
}

type MusicAlbum struct {
	*ItemInfo
	AlbumArtist string
	Artist      []string
	ArtistThumb string
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
