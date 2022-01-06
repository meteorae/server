package utils

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func HashFilePath(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := h.Write([]byte(path)); err != nil {
		return nil, fmt.Errorf("could not write to hash: %w", err)
	}

	return h.Sum(nil), nil
}

func HashFileBytes(fileBytes []byte) ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write(fileBytes); err != nil {
		return nil, fmt.Errorf("could not write to hash: %w", err)
	}

	return h.Sum(nil), nil
}
