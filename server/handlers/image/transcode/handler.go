package transcode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gorilla/schema"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"
)

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
		magickWand := imagick.NewMagickWand()
		defer magickWand.Destroy()

		var buffer bytes.Buffer

		// If it's an external URL, we can download it, check if it's an image, and then transcode it
		// TODO: We probably want a whitelist of domains, to cases where an attacker would request tons of large files to cause a DOS
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
				http.Error(writer, "Internal URLs are not handled yet", http.StatusBadRequest)
			}

			metadataID := match[1]
			metadataImageType := match[2]

			// Check if it's a valid type, mainly to avoid people querying the API throuth this endpoint
			if metadataImageType != "thumb" && metadataImageType != "art" {
				http.Error(writer, "Unauthorized image type", http.StatusBadRequest)

				return
			}

			var metadata models.ItemMetadata

			result := database.DB.Select(metadataImageType).Find(&metadata, "id = ?", metadataID)
			if result.Error != nil {
				http.Error(writer, "Item not found", http.StatusInternalServerError)

				return
			}

			if metadataImageType == "thumb" && metadata.Thumb == "" || metadataImageType == "art" && metadata.Art == "" {
				http.Error(writer, "Image not found", http.StatusInternalServerError)

				return
			}

			if metadataImageType == "thumb" {
				imageHash = metadata.Thumb
			} else if metadataImageType == "art" {
				imageHash = metadata.Art
			}

			prefix := imageHash[0:2]

			baseDirectory := filepath.Join(handler.imageCachePath, prefix, imageHash)

			var imagePath string

			if imageQuery.Height != 0 && imageQuery.Width != 0 {
				imagePath = filepath.Join(baseDirectory, fmt.Sprintf("%dx%d.webp", imageQuery.Width, imageQuery.Height))

				if _, err := os.Stat(imagePath); err == nil {
					isCached = true
				} else if errors.Is(err, os.ErrNotExist) {
					isCached = false
					shouldCache = true

					// Set this back to the original, since we need to generate the image with the proper size
					imagePath = filepath.Join(baseDirectory, "0x0.webp")
				} else {
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

		err = magickWand.ReadImageBlob(buffer.Bytes())
		if err != nil {
			log.Err(err).Msg("Failed to read image into MagickWand")
			http.Error(writer, "Failed to read image into MagickWand", http.StatusInternalServerError)

			return
		}

		// If the request wants a specific size, resize the image
		if imageQuery.Height != 0 && imageQuery.Width != 0 {
			sourceWidth := magickWand.GetImageWidth()
			sourceHeight := magickWand.GetImageHeight()

			widthRatio := float64(imageQuery.Width) / float64(sourceWidth)
			heightRatio := float64(imageQuery.Height) / float64(sourceHeight)
			bestRatio := math.Min(widthRatio, heightRatio)

			newWidth := uint(float64(sourceWidth) * bestRatio)
			newHeight := uint(float64(sourceHeight) * bestRatio)

			err = magickWand.ResizeImage(newWidth, newHeight, imagick.FILTER_LANCZOS2_SHARP)
			if err != nil {
				log.Err(err).Msg("Failed to resize image")
				http.Error(writer, "Failed to resize image", http.StatusInternalServerError)

				return
			}

			if shouldCache {
				prefix := imageHash[0:2]
				baseDirectory := filepath.Join(handler.imageCachePath, prefix, imageHash)
				imagePath := filepath.Join(baseDirectory, fmt.Sprintf("%dx%d.webp", imageQuery.Width, imageQuery.Height))

				err = magickWand.SetFormat("webp")
				if err != nil {
					log.Err(err).Msg("Failed to set image format")
					http.Error(writer, "Failed to set image format", http.StatusInternalServerError)
				}

				err = magickWand.WriteImage(imagePath)
				if err != nil {
					log.Err(err).Msg("Failed to write image")
					http.Error(writer, "Failed to write image", http.StatusInternalServerError)

					return
				}

				isCached = true
			}
		}

		// We store images internally in WebP for size and quality reasons.
		// If the client doesn't support it, we convert it to JPEG
		if isCached && !supportsWebP(acceptList) {
			err = magickWand.SetImageFormat("jpg")
			if err != nil {
				log.Err(err).Msg("Failed to set image format")
				http.Error(writer, "Failed to set image format", http.StatusInternalServerError)

				return
			}
		}

		buffer = *bytes.NewBuffer(magickWand.GetImageBlob())

		filetype := http.DetectContentType(buffer.Bytes())

		writer.Header().Set("Content-Type", filetype)

		_, err = writer.Write(buffer.Bytes())
		if err != nil {
			log.Err(err).Msg("Failed to write image")
			http.Error(writer, "Failed to write image", http.StatusInternalServerError)

			return
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
