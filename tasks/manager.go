package tasks

import (
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"github.com/vmihailenco/taskq/v3/memqueue"
)

var QueueFactory = memqueue.NewFactory()

var LibraryScanQueue *ants.Pool

var MetadataRefreshQueue *ants.Pool

var MediaAnalysisQueue *ants.Pool

func StartTaskQueues() error {
	log.Info().Msg("Starting task queues")

	var err error

	LibraryScanQueue, err = ants.NewPool(1)
	if err != nil {
		return err
	}

	MetadataRefreshQueue, err = ants.NewPool(2)
	if err != nil {
		return err
	}

	MediaAnalysisQueue, err = ants.NewPool(2)
	if err != nil {
		return err
	}

	return nil
}

func StopTaskQueues() {
	log.Info().Msg("Stopping task queues")

	LibraryScanQueue.Release()
	MetadataRefreshQueue.Release()
	MediaAnalysisQueue.Release()
}
