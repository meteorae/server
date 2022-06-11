package transcode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gorilla/schema"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/rs/zerolog/log"
)

type ImageType string

const (
	ThumbImage ImageType = "thumb"
	ArtImage   ImageType = "art"
)

func (t ImageType) String() string {
	return string(t)
}

type ImageQuery struct {
	URL    string `schema:"url,required"`
	Width  int    `schema:"width"`
	Height int    `schema:"height"`
}

type ImageHandler struct {
	imageCachePath string
}

var validInternalURLRegexp = regexp.MustCompile(`^\/metadata\/(\d*)\/([a-z]*)$`)

func NewImageHandler() (*ImageHandler, error) {
	imageCachePath, err := xdg.CacheFile("meteorae/images")
	if err != nil {
		return nil, fmt.Errorf("failed to get image cache path: %w", err)
	}

	return &ImageHandler{
		imageCachePath: imageCachePath,
	}, nil
}

func (handler *ImageHandler) HTTPHandler(writer http.ResponseWriter, request *http.Request) {
	// Require authentication, to avoid non-users slamming the API with external image requests
	/*user := utils.GetUserFromContext(request.Context())
	if user == nil {
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
	}*/

	isCached := false
	shouldCache := false

	var imageHash string

	var imageQuery ImageQuery
	if err := schema.NewDecoder().Decode(&imageQuery, request.URL.Query()); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)

		return
	}

	accept := request.Header.Get("Accept")
	acceptList := strings.Split(accept, ",")

	if imageQuery.URL != "" {
		var buffer bytes.Buffer

		// If it's an external URL, we can download it, check if it's an image, and then transcode it
		parsedURL, err := url.ParseRequestURI(imageQuery.URL)
		if err != nil {
			log.Err(err).Msg("Failed to parse URL")
			http.Error(writer, "Invalid URL", http.StatusBadRequest)

			return
		}

		if strings.HasPrefix(imageQuery.URL, "/") {
			match := validInternalURLRegexp.FindStringSubmatch(imageQuery.URL)
			if match == nil {
				// This is a malformed URL
				http.Error(writer, "URL is malformed", http.StatusBadRequest)
			}

			metadataID := match[1]
			metadataImageType := match[2]

			// Check if it's a valid type, mainly to avoid people querying the API throuth this endpoint
			if metadataImageType != ThumbImage.String() && metadataImageType != ArtImage.String() {
				http.Error(writer, "Unauthorized image type", http.StatusBadRequest)

				return
			}

			metadata, err := database.GetItemByID(metadataID)
			if err != nil {
				log.Err(err).Msg("Failed to get metadata")

				return
			}

			if metadataImageType == ThumbImage.String() && metadata.Thumb == "" || metadataImageType == ArtImage.String() && metadata.Art == "" {
				http.Error(writer, "Image not found", http.StatusInternalServerError)

				return
			}

			if metadataImageType == ThumbImage.String() {
				imageHash = metadata.Thumb
			} else if metadataImageType == ArtImage.String() {
				imageHash = metadata.Art
			}

			prefix := imageHash[0:2]

			baseDirectory := filepath.Join(handler.imageCachePath, prefix, imageHash)

			var imagePath string

			if imageQuery.Height != 0 && imageQuery.Width != 0 {
				imagePath = filepath.Join(baseDirectory, fmt.Sprintf("%dx%d.webp", imageQuery.Width, imageQuery.Height))

				_, err := os.Stat(imagePath)

				switch {
				case errors.Is(err, nil):
					isCached = true
				case errors.Is(err, os.ErrNotExist):
					isCached = false
					shouldCache = true

					// Set this back to the original, since we need to generate the image with the proper size
					imagePath = filepath.Join(baseDirectory, "0x0.webp")
				default:
					http.Error(writer, "Failed to stat image", http.StatusInternalServerError)

					return
				}
			} else {
				// 0x0 means we want the original image
				imagePath = filepath.Join(baseDirectory, "0x0.webp")

				// Defaut images are always cached. If it's not, there's a problem
				isCached = true
			}

			data, err := os.ReadFile(imagePath)
			if err != nil {
				http.Error(writer, "Failed to read image", http.StatusInternalServerError)

				return
			}

			buffer = *bytes.NewBuffer(data)
		} else {
			// We don't cache external images, so we don't need to check if it's already cached
			response, err := http.Get(parsedURL.String())
			if err != nil {
				log.Err(err).Msg("Failed to download image")
				http.Error(writer, "Failed to download image", http.StatusInternalServerError)

				return
			}
			defer response.Body.Close()

			_, err = io.Copy(&buffer, response.Body)
			if err != nil {
				log.Err(err).Msg("Failed to read image")
				http.Error(writer, "Failed to read image", http.StatusInternalServerError)

				return
			}
		}

		image, err := vips.NewImageFromReader(&buffer)
		if err != nil {
			log.Err(err).Msg("Failed to read image into VIPS")
			http.Error(writer, "Failed to read image into VIPS", http.StatusInternalServerError)

			return
		}

		// If the request wants a specific size, resize the image
		if imageQuery.Height != 0 && imageQuery.Width != 0 {
			sourceWidth := image.Width()
			sourceHeight := image.Height()

			widthRatio := float64(imageQuery.Width) / float64(sourceWidth)
			heightRatio := float64(imageQuery.Height) / float64(sourceHeight)
			bestRatio := math.Min(widthRatio, heightRatio)

			newWidth := float64(sourceWidth) * bestRatio
			newHeight := float64(sourceHeight) * bestRatio

			err = image.Thumbnail(int(newWidth), int(newHeight), vips.InterestingNone)
			if err != nil {
				log.Err(err).Msg("Failed to resize image")
				http.Error(writer, "Failed to resize image", http.StatusInternalServerError)

				return
			}

			if shouldCache {
				prefix := imageHash[0:2]
				baseDirectory := filepath.Join(handler.imageCachePath, prefix, imageHash)
				imagePath := filepath.Join(baseDirectory, fmt.Sprintf("%dx%d.webp", imageQuery.Width, imageQuery.Height))

				export, _, err := image.ExportWebp(vips.NewWebpExportParams())
				if err != nil {
					log.Err(err).Msg("Failed to set image format")
					http.Error(writer, "Failed to set image format", http.StatusInternalServerError)
				}

				err = ioutil.WriteFile(imagePath, export, helpers.BaseFilePermissions)
				if err != nil {
					log.Err(err).Msg("Failed to write image")
					http.Error(writer, "Failed to write image", http.StatusInternalServerError)

					return
				}

				isCached = true

				buffer = *bytes.NewBuffer(export)
			}
		}

		// We store images internally in WebP for size and quality reasons.
		// If the client doesn't support it, we convert it to JPEG
		if isCached && !supportsWebP(acceptList) {
			export, _, err := image.ExportJpeg(vips.NewJpegExportParams())
			if err != nil {
				log.Err(err).Msg("Failed to set image format")
				http.Error(writer, "Failed to set image format", http.StatusInternalServerError)

				return
			}

			buffer = *bytes.NewBuffer(export)
		}

		export, _, err := image.ExportJpeg(vips.NewJpegExportParams())
		if err != nil {
			log.Err(err).Msg("Failed to set image format")
			http.Error(writer, "Failed to set image format", http.StatusInternalServerError)

			return
		}

		buffer = *bytes.NewBuffer(export)

		writer.Header().Set("Cache-Control", "max-age=604800")

		_, err = io.Copy(writer, &buffer)
		if err != nil {
			log.Err(err).Msg("Failed to write image")
			http.Error(writer, "Failed to write image", http.StatusInternalServerError)
		}
	} else {
		http.Error(writer, "Invalid URL", http.StatusBadRequest)

		return
	}
}

func supportsWebP(acceptList []string) bool {
	for _, accept := range acceptList {
		if accept == "image/webp" {
			return true
		}
	}

	return false
}
