// Package react is a handler for the olaris-react application.
package web

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed client/*
var embedded embed.FS

// SPAHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type SPAHandler struct{}

func (h SPAHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {
	serverRoot, fsErr := fs.Sub(embedded, ".")
	if fsErr != nil {
		log.Error().Err(fsErr).Msg("Failed to get server root")
		http.Error(writer, "Failed to get server root", http.StatusInternalServerError)
	}

	path, absErr := filepath.Abs(reader.URL.Path)
	if absErr != nil {
		log.Error().Err(absErr).Msg("Failed to get absolute path")
		http.Error(writer, absErr.Error(), http.StatusBadRequest)

		return
	}

	log.Debug().Str("path", filepath.Join("client", path)).Msg("Serving file")

	fileStat, fsStatErr := fs.Stat(serverRoot, filepath.Join("client", path))
	if os.IsNotExist(fsStatErr) || fileStat.IsDir() && path == "/" {
		indexFile, indexOpenErr := serverRoot.Open(filepath.Join("client", "index.html"))
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

		http.ServeContent(writer, reader, "index.html", indexStat.ModTime(), indexFile.(io.ReadSeeker))

		return
	} else if fsStatErr != nil {
		log.Error().Err(fsStatErr).Msg("Failed to stat file")
		http.Error(writer, fsStatErr.Error(), http.StatusInternalServerError)

		return
	}

	spaFile, fsFileReadErr := fs.ReadFile(serverRoot, filepath.Join("client", path))
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
