package database

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func SetAcoustID(mediaPart *MediaPart, acoustID string) {
	mediaPart.AcoustID = acoustID

	db.Save(&mediaPart)
}

func CreateMediaStream(title string, streamType StreamType, language string, index int,
	streamInfo datatypes.JSON, mediaPartID uint64) error {
	mediaStream := MediaStream{
		Title:           title,
		StreamType:      streamType,
		Language:        language,
		Index:           index,
		MediaStreamInfo: streamInfo,
		MediaPartID:     mediaPartID,
	}

	result := db.Create(&mediaStream)
	if result.Error != nil {
		log.Err(result.Error).Msgf("Could not create stream")

		return result.Error
	}

	return nil
}

func CreateMediaPart(path, hash string, size int64) (*MediaPart, error) {
	newMediaPart := MediaPart{
		FilePath: path,
		Hash:     hash,
		Size:     size,
	}

	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newMediaPart)
	// TODO: Check for the actual error type
	if result.Error != nil {
		// If the record exist, we already have it, just skip it to save time
		return nil, fmt.Errorf("failed to create media part: %w", result.Error)
	}

	return &newMediaPart, nil
}

func GetMediaPart(metadataID, mediaPartID string) (*MediaPart, error) {
	var mediaPart MediaPart

	result := db.Find(&mediaPart, "item_metadata_id = ? AND id = ?", metadataID, mediaPartID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mediaPart, nil
}
