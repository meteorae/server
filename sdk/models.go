// Package sdk provides common structs and plugin facilities for Meteorae.
package sdk

import (
	"encoding/xml"
	"time"

	"github.com/google/uuid"
)

type ItemType uint

func (t ItemType) String() string {
	switch t {
	case MovieItem:
		return "Movie"
	case MusicAlbumItem:
		return "MusicAlbum"
	case MusicMediumItem:
		return "MusicMedium"
	case MusicTrackItem:
		return "MusicTrack"
	case TVSeasonItem:
		return "TVSeason"
	case TVShowItem:
		return "TVShow"
	case TVEpisodeItem:
		return "TVEpisode"
	case ImageItem:
		return "Image"
	case ImageAlbumItem:
		return "ImageAlbum"
	case PersonItem:
		return "Person"
	case CollectionItem:
		return "Collection"
	case VideoClipItem:
		return "VideoClip"
	case UnknownItem:
		return "Unknown"
	}

	return "Unknown"
}

const (
	MovieItem ItemType = iota
	MusicAlbumItem
	MusicMediumItem
	MusicTrackItem
	TVShowItem
	TVSeasonItem
	TVEpisodeItem
	ImageItem
	ImageAlbumItem
	PersonItem
	CollectionItem
	VideoClipItem
	UnknownItem
)

type Item interface {
	IsItem()
	GetID() uint
	GetTitle() string
	GetReleaseDate() time.Time
	GetUUID() uuid.UUID
	GetIdentifiers() []Identifier
	GetThumbs() []ItemImage
	GetArt() []ItemImage
	GetType() ItemType
	GetParts() []string
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
	ID            uint         `xml:"id,attr"`
	UUID          uuid.UUID    `xml:"uuid,attr"`
	Title         string       `xml:"title"`
	SortTitle     string       `xml:"sort_title"`
	OriginalTitle string       `xml:"original_title"`
	ReleaseDate   time.Time    `xml:"release_date" json:"startDate"`
	Parts         []string     `xml:"-"`
	Thumb         Posters      `xml:"posters"`
	Art           Art          `xml:"art"`
	Language      string       `xml:"language"`
	Identifiers   []Identifier `xml:"identifier"`
	MatchScore    uint         `xml:"-"`
	LibraryID     uint         `xml:"-"`
	CreatedAt     time.Time    `xml:"-"`
	UpdatedAt     time.Time    `xml:"-"`
	DeletedAt     time.Time    `xml:"-"`
	IsRefreshing  bool         `xml:"-" json:"isRefreshing"`
	IsAnalyzing   bool         `xml:"-" json:"isAnalyzing"`
}

func (i ItemInfo) IsItem() {}

func (i ItemInfo) GetID() uint {
	return i.ID
}

func (i ItemInfo) GetTitle() string {
	return i.Title
}

func (i ItemInfo) GetReleaseDate() time.Time {
	return i.ReleaseDate
}

func (i ItemInfo) GetUUID() uuid.UUID {
	return i.UUID
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

func (i ItemInfo) GetType() ItemType {
	return UnknownItem
}

func (i ItemInfo) GetParts() []string {
	return i.Parts
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
	IdentifierType IdentifierType `xml:"type,attr" json:"type"`
	Identifier     string         `xml:"value,attr" json:"value"`
}

// A Movie represents information about an individual movie, obtained through scanning or through an agent.
type Movie struct {
	XMLName xml.Name `xml:"movie"`
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
}

func (m Movie) GetType() ItemType {
	return MovieItem
}

// A TVShow represents information about a TV series, obtained through scanning or through an agent.
type TVShow struct {
	XMLName xml.Name `xml:"tvshow"`
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
}

func (t TVShow) GetType() ItemType {
	return TVShowItem
}

type TVSeason struct {
	XMLName xml.Name `xml:"season"`
	*ItemInfo
	Summary     string
	SeriesTitle string
}

func (t TVSeason) GetType() ItemType {
	return TVSeasonItem
}

// A TVEpisode represents information about an individual episode of a TV series,
// obtained through scanning or through an agent.
type TVEpisode struct {
	XMLName xml.Name `xml:"tvepisode"`
	*ItemInfo
	Tagline    string
	Summary    string
	Genres     []string
	Popularity float32
	Studios    []string
	Countries  []string
	// Credits       []Credit
	SeriesTitle string
	SeriesID    uint
	SeasonTitle string
	Season      int
	SeasonID    uint
	Episode     int
}

func (t TVEpisode) GetType() ItemType {
	return TVEpisodeItem
}

type MusicAlbum struct {
	XMLName xml.Name `xml:"album"`
	*ItemInfo
	AlbumArtist string
	Artist      []string
	ArtistThumb string
}

func (m MusicAlbum) GetType() ItemType {
	return MusicAlbumItem
}

type MusicMedium struct {
	XMLName xml.Name `xml:"medium"`
	*ItemInfo
	Title       string
	AlbumArtist string
}

// A MusicTrack represents information about an individual track of a music album,
// obtained through scanning or through an agent.
type MusicTrack struct {
	XMLName xml.Name `xml:"track"`
	*ItemInfo
	AlbumArtist string
	AlbumName   string
	DiscIndex   int
	TrackIndex  int
	Artist      []string
}

func (m MusicTrack) GetType() ItemType {
	return MusicTrackItem
}

type Image struct {
	XMLName xml.Name `xml:"image"`
	*ItemInfo
}

func (i Image) GetType() ItemType {
	return ImageItem
}

type Video struct {
	XMLName xml.Name `xml:"video"`
	*ItemInfo
}

func (i Video) GetType() ItemType {
	return VideoClipItem
}
