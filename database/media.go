package database

import (
	"fmt"
	"time"

	"github.com/ostafen/clover"
	"gopkg.in/vansante/go-ffprobe.v2"
	"gorm.io/datatypes"
)

type MediaPart struct {
	Id               uint64        `clover:"_id"`
	Hash             string        `clover:"hash"`
	OpenSubtitleHash string        `clover:"openSubtitleHash"`
	AniDBCRC         string        `clover:"aniDbCRC"`
	AcoustID         string        `clover:"acoustId"`
	FilePath         string        `clover:"filePath"`
	Size             int64         `clover:"size"`
	ItemMetadataID   uint64        `clover:"itemMetadataId"`
	MediaStreams     []MediaStream `clover:"mediaStreams"`
	CreatedAt        time.Time     `clover:"createdAt"`
	UpdatedAt        time.Time     `clover:"updatedAt"`
	DeletedAt        time.Time     `clover:"deletedAt"`
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

func (t IdentifierType) String() string {
	return [...]string{
		"IMDB ID",
		"TheMovieDB ID",
		"AniDB ID",
		"TVDB ID",
		"MusicBrainz ID",
		"Facebook ID",
		"Twitter ID",
		"Instagram ID",
	}[t]
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

func SetAcoustID(mediaPart *MediaPart, acoustId string) {
	updates := make(map[string]interface{})
	updates["acoustId"] = acoustId
	updates["updatedAt"] = time.Now()

	query := db.Query(MediaPartCollection.String()).Where(clover.Field("_id").Eq(mediaPart.Id))
	query.Update(updates)
}

func CreateMediaStream(title string, streamType StreamType, language string, index int,
	streamInfo datatypes.JSON, mediaPartID uint64,
) error {
	mediaStream := MediaStream{
		Title:           title,
		StreamType:      streamType,
		Language:        language,
		Index:           index,
		MediaStreamInfo: streamInfo,
		MediaPartID:     mediaPartID,
	}

	document := clover.NewDocumentOf(&mediaStream)

	if _, err := db.InsertOne(MediaStreamCollection.String(), document); err != nil {
		return fmt.Errorf("failed to create media stram: %w", err)
	}

	return nil
}

func CreateMediaPart(mediaPart MediaPart) (*MediaPart, error) {
	document := clover.NewDocumentOf(&mediaPart)

	if _, err := db.InsertOne(MediaPartCollection.String(), document); err != nil {
		return nil, fmt.Errorf("failed to create media part: %w", err)
	}

	return &mediaPart, nil
}

func GetMediaPartById(metadataID, mediaPartID string) (*MediaPart, error) {
	var mediaPart MediaPart

	mediaPartDocument, err := db.Query(LibraryCollection.String()).Where(clover.Field("_id").Eq(mediaPartID).
		And(clover.Field("item_metadata_id").Eq(metadataID))).FindFirst()
	if err != nil {
		return &mediaPart, fmt.Errorf("failed to get media part: %w", err)
	}

	mediaPartDocument.Unmarshal(&mediaPart)

	return &mediaPart, nil
}
