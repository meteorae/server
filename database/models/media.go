package models

import (
	"time"

	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MediaPart struct {
	ID               uint64 `gorm:"primary_key" json:"id"`
	Hash             string `gorm:"not null"`
	OpenSubtitleHash string
	AniDBCRC         string
	AcoustID         string
	FilePath         string `gorm:"unique;not null"`
	Size             int64  `gorm:"not null"`
	ItemMetadataID   uint64
	MediaStreams     []MediaStream  `json:"mediaStreams"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type IdentifierType int8

const (
	ImdbIdentifier IdentifierType = iota
	TmdbIdentifier
	AnidbIdentifier
	TvdbIdentifier
	MusicbrainzIdentifier
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
		"MusicBrainz ID",
		"Facebook ID",
		"Twitter ID",
		"Instagram ID",
	}[d]
}

type ExternalIdentifier struct {
	ID             uint64         `gorm:"primary_key" json:"id"`
	IdentifierType IdentifierType `gorm:"not null"`
	Identifier     string         `gorm:"not null"`
	MovieID        uint64         `gorm:"not null"`
}

type StreamType int8

const (
	VideoStream StreamType = iota
	AudioStream
	SubtitleStream
)

func (d StreamType) String() string {
	return [...]string{"Video", "Audio", "Subtitle"}[d]
}

type MediaStream struct {
	ID         uint64     `gorm:"primary_key" json:"id"`
	Title      string     `json:"title"`
	StreamType StreamType `gorm:"not null"`
	Language   string     `json:"language"`
	Index      int        `gorm:"not null"`
	// This is technically a MediaStreamInfo
	MediaStreamInfo datatypes.JSON `json:"mediaStreamInfo"`
	MediaPartID     uint64         `gorm:"not null"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

// This is stored in DB as a JSON blob, since we never need to filter on it.
// All properties are optional, and more can be added later.
type MediaStreamInfo struct {
	CodecName          string                    `json:"codecName"`
	CodecLongName      string                    `json:"codecLongName"`
	CodecType          string                    `json:"codecType"`
	CodecTimeBase      string                    `json:"codecTimeBase"`
	CodecTag           string                    `json:"codecTag"`
	RFrameRate         string                    `json:"rFrameRate"`
	AvgFrameRate       string                    `json:"avgFrameRate"`
	TimeBase           string                    `json:"timeBase"`
	StartPts           int                       `json:"startPts"`
	StartTime          string                    `json:"startTime"`
	DurationTS         uint64                    `json:"durationTs"`
	Duration           string                    `json:"duration"`
	BitRate            string                    `json:"bitrate"`
	BitsPerRawSample   string                    `json:"bitsPerRawSample"`
	NbFrames           string                    `json:"nbFrames"`
	Disposition        ffprobe.StreamDisposition `json:"disposition"`
	Tags               ffprobe.StreamTags        `json:"tags"`
	Profile            string                    `json:"profile"`
	Width              int                       `json:"width"`
	Height             int                       `json:"height"`
	HasBFrames         int                       `json:"hasBFrames"`
	SampleAspectRatio  string                    `json:"sampleAspectRatio"`
	DisplayAspectRatio string                    `json:"displayAspectRatio"`
	PixelFormat        string                    `json:"pixelFormat"`
	Level              int                       `json:"level"`
	ColorRange         string                    `json:"colorRange"`
	ColorSpace         string                    `json:"colorSpace"`
	SampleFmt          string                    `json:"sampleFmt"`
	SampleRate         string                    `json:"sampleRate"`
	Channels           int                       `json:"channels"`
	ChannelsLayout     string                    `json:"channelLayout"`
	BitsPerSample      int                       `json:"bitsPerSample"`
}
