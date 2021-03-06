// Package react is a handler for the olaris-react application.
package web

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/mholt/archiver/v3"
	"github.com/rs/zerolog/log"
)

func EnsureWebClient() error {
	webAssetsLocation, dataFileErr := xdg.DataFile("meteorae/assets")
	if dataFileErr != nil {
		return fmt.Errorf("could not get path for web assets: %w", dataFileErr)
	}

	serverRoot := os.DirFS(webAssetsLocation)

	webClientStat, fsStatErr := fs.Stat(serverRoot, "web")

	switch {
	case os.IsNotExist(fsStatErr):
		log.Debug().Msg("Web client not found, downloading from GitHub")

		// TODO: Ideally we should go through Github's API to get the version as well
		resp, err := http.Get("https://github.com/meteorae/web/releases/latest/download/web.zip")
		if err != nil {
			return fmt.Errorf("failed to download web client: %w", err)
		}
		defer resp.Body.Close()

		webClientZipFile, err := os.CreateTemp("", "meteorae-web-client-*.zip")
		if err != nil {
			return fmt.Errorf("failed to create temp file for web client: %w", err)
		}

		_, err = io.Copy(webClientZipFile, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to copy web client archive: %w", err)
		}

		zipHandler := archiver.NewZip()

		err = zipHandler.Unarchive(webClientZipFile.Name(), filepath.Join(webAssetsLocation, "web"))
		if err != nil {
			return fmt.Errorf("failed to extract web client: %w", err)
		}

		log.Debug().Msg("Web client downloaded successfully")
	case webClientStat.IsDir():
		log.Debug().Msg("Web client already exists")

		return nil
	default:
		return fmt.Errorf("could not stat web client: %w", fsStatErr)
	}

	return nil
}

// SPAHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SPAHandler struct{}

func (h SPAHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {
	webAssetsLocation, dataFileErr := xdg.DataFile("meteorae/assets")
	if dataFileErr != nil {
		log.Error().Err(dataFileErr).Msg("could not get path for web assets")
		http.Error(writer, "Failed to get web assets directory", http.StatusInternalServerError)
	}

	serverRoot := os.DirFS(webAssetsLocation)

	path, absErr := filepath.Abs(reader.URL.Path)
	if absErr != nil {
		log.Error().Err(absErr).Msg("Failed to get absolute path")
		http.Error(writer, absErr.Error(), http.StatusBadRequest)

		return
	}

	log.Debug().Str("path", filepath.Join("web", path)).Msg("Serving file")

	fileStat, fsStatErr := fs.Stat(serverRoot, filepath.Join("web", path))
	if os.IsNotExist(fsStatErr) || fileStat.IsDir() && path == "/" {
		indexFile, indexOpenErr := serverRoot.Open(filepath.Join("web", "index.html"))
		if indexOpenErr != nil {
			log.Error().Err(indexOpenErr).Msg("Failed to open index.html")
			http.Error(writer, indexOpenErr.Error(), http.StatusInternalServerError)

			return
		}
		defer indexFile.Close()

		indexStat, indexStatErr := indexFile.Stat()
		if indexStatErr != nil {
			log.Error().Err(indexStatErr).Msg("Failed to stat index.html")
			http.Error(writer, indexStatErr.Error(), http.StatusForbidden)

			return
		}

		indexFileSeeker, ok := indexFile.(io.ReadSeeker)
		if !ok {
			log.Error().Msg("Failed to cast index.html to io.ReadSeeker")
			http.Error(writer, "Failed to cast index.html to io.ReadSeeker", http.StatusInternalServerError)
		}

		http.ServeContent(writer, reader, "index.html", indexStat.ModTime(), indexFileSeeker)

		return
	} else if fsStatErr != nil {
		log.Error().Err(fsStatErr).Msg("Failed to stat file")
		http.Error(writer, fsStatErr.Error(), http.StatusInternalServerError)

		return
	}

	spaFile, fsFileReadErr := fs.ReadFile(serverRoot, filepath.Join("web", path))
	if fsFileReadErr != nil {
		log.Error().Err(fsFileReadErr).Msg("Failed to open index.html")
		http.Error(writer, fsFileReadErr.Error(), http.StatusInternalServerError)

		return
	}

	if strings.HasSuffix(path, ".js") {
		writer.Header().Set("Content-Type", "application/javascript")
	} else {
		contentType := http.DetectContentType(spaFile)
		writer.Header().Set("Content-Type", contentType)
	}

	spaFileBuffer := bytes.NewReader(spaFile)

	http.ServeContent(writer, reader, path, fileStat.ModTime(), spaFileBuffer)
}
