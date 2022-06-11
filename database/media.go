package database

import (
	"fmt"
	"time"

	"github.com/ostafen/clover"
	"gopkg.in/vansante/go-ffprobe.v2"
)

type MediaPart struct {
	Id               string        `clover:"_id" json:"_id"` //nolint:tagliatelle
	Hash             string        `clover:"hash"`
	OpenSubtitleHash string        `clover:"openSubtitleHash"`
	AniDBCRC         string        `clover:"aniDbCRC"`
	AcoustID         string        `clover:"acoustId"`
	Path             string        `clover:"path"`
	Size             int64         `clover:"size"`
	ItemId           string        `clover:"itemId"`
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
	Title       string `clover:"title"`
	Type        StreamType
	Language    string `clover:"language"`
	Index       int
	Information MediaStreamInfo `clover:"infoformation"`
	CreatedAt   time.Time       `clover:"createdAt"`
	UpdatedAt   time.Time       `clover:"updatedAt"`
}

// This is stored in DB as a JSON blob, since we never need to filter on it.
// All properties are optional, and more can be added later.
type MediaStreamInfo struct {
	CodecName          string                    `clover:"codecName"`
	CodecLongName      string                    `clover:"codecLongName"`
	CodecType          string                    `clover:"codecType"`
	CodecTimeBase      string                    `clover:"codecTimeBase"`
	CodecTag           string                    `clover:"codecTag"`
	RFrameRate         string                    `clover:"rFrameRate"`
	AvgFrameRate       string                    `clover:"avgFrameRate"`
	TimeBase           string                    `clover:"timeBase"`
	StartPts           int                       `clover:"startPts"`
	StartTime          string                    `clover:"startTime"`
	DurationTS         uint64                    `clover:"durationTs"`
	Duration           string                    `clover:"duration"`
	BitRate            string                    `clover:"bitrate"`
	BitsPerRawSample   string                    `clover:"bitsPerRawSample"`
	NbFrames           string                    `clover:"nbFrames"`
	Disposition        ffprobe.StreamDisposition `clover:"disposition"`
	Tags               ffprobe.StreamTags        `clover:"tags"`
	Profile            string                    `clover:"profile"`
	Width              int                       `clover:"width"`
	Height             int                       `clover:"height"`
	HasBFrames         int                       `clover:"hasBFrames"`
	SampleAspectRatio  string                    `clover:"sampleAspectRatio"`
	DisplayAspectRatio string                    `clover:"displayAspectRatio"`
	PixelFormat        string                    `clover:"pixelFormat"`
	Level              int                       `clover:"level"`
	ColorRange         string                    `clover:"colorRange"`
	ColorSpace         string                    `clover:"colorSpace"`
	SampleFmt          string                    `clover:"sampleFmt"`
	SampleRate         string                    `clover:"sampleRate"`
	Channels           int                       `clover:"channels"`
	ChannelsLayout     string                    `clover:"channelLayout"`
	BitsPerSample      int                       `clover:"bitsPerSample"`
}

func SetAcoustID(mediaPart *MediaPart, acoustId string) error {
	updates := make(map[string]interface{})
	updates["acoustId"] = acoustId
	updates["updatedAt"] = time.Now()

	query := db.Query(MediaPartCollection.String()).Where(clover.Field("_id").Eq(mediaPart.Id))

	if err := query.Update(updates); err != nil {
		return err
	}

	return nil
}

func CreateMediaStream(title string, streamType StreamType, language string, index int,
	streamInfo MediaStreamInfo, mediaPart MediaPart,
) error {
	mediaStream := MediaStream{
		Title:       title,
		Type:        streamType,
		Language:    language,
		Index:       index,
		Information: streamInfo,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	existingStreams := mediaPart.MediaStreams

	existingStreams = append(existingStreams, mediaStream)

	updates := make(map[string]interface{})
	updates["mediaStreams"] = existingStreams

	query := db.Query(ItemCollection.String()).Where(clover.Field("_id").Eq(mediaPart.Id))
	query.Update(updates)

	return nil
}

func CreateMediaPart(mediaPart MediaPart) (*MediaPart, error) {
	mediaPart.CreatedAt = time.Now()
	mediaPart.UpdatedAt = time.Now()

	document := clover.NewDocumentOf(&mediaPart)

	mediaPartId, err := db.InsertOne(MediaPartCollection.String(), document)
	if err != nil {
		return nil, fmt.Errorf("failed to create media part: %w", err)
	}

	mediaPart.Id = mediaPartId

	return &mediaPart, nil
}

func GetMediaPartById(metadataID, mediaPartID string) (*MediaPart, error) {
	var mediaPart MediaPart

	mediaPartDocument, err := db.Query(LibraryCollection.String()).Where(clover.Field("_id").Eq(mediaPartID).
		And(clover.Field("itemId").Eq(metadataID))).FindFirst()
	if err != nil {
		return &mediaPart, fmt.Errorf("failed to get media part: %w", err)
	}

	mediaPartDocument.Unmarshal(&mediaPart)

	return &mediaPart, nil
}

func GetMediaPartByItemId(metadataID string) (*MediaPart, error) {
	var mediaPart MediaPart

	mediaPartDocument, err := db.Query(MediaPartCollection.String()).Where(clover.Field("itemId").
		Eq(metadataID)).FindFirst()
	if err != nil {
		return &mediaPart, fmt.Errorf("failed to get media part: %w", err)
	}

	mediaPartDocument.Unmarshal(&mediaPart)

	return &mediaPart, nil
}
