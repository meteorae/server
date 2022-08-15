package tasks

import (
	"fmt"

	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

const (
	libraryScanQueueSize     = 1
	metadataRefreshQueueSize = 2
	mediaAnalyzisQueueSize   = 2
)

var (
	LibraryScanQueue     *ants.Pool //nolint:gochecknoglobals // Required for now
	MetadataRefreshQueue *ants.Pool //nolint:gochecknoglobals // Required for now
	MediaAnalysisQueue   *ants.Pool //nolint:gochecknoglobals // Required for now
)

func StartTaskQueues() error {
	log.Info().Msg("Starting task queues")

	var err error

	LibraryScanQueue, err = ants.NewPool(libraryScanQueueSize)
	if err != nil {
		return fmt.Errorf("failed to create library scan queue: %w", err)
	}

	MetadataRefreshQueue, err = ants.NewPool(metadataRefreshQueueSize)
	if err != nil {
		return fmt.Errorf("failed to create metadata refresh queue: %w", err)
	}

	MediaAnalysisQueue, err = ants.NewPool(mediaAnalyzisQueueSize)
	if err != nil {
		return fmt.Errorf("failed to create media analysis queue: %w", err)
	}

	return nil
}

func StopTaskQueues() {
	log.Info().Msg("Stopping task queues")

	LibraryScanQueue.Release()
	MetadataRefreshQueue.Release()
	MediaAnalysisQueue.Release()
}
