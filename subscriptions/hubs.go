package subscriptions

import (
	"sync"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

type hubUpdatedSubscription struct {
	librariesUpdated map[uint]database.Library
	timer            *time.Timer
	mu               sync.Mutex
}

var (
	itemsAdded         []*models.Item
	hubUpdatedNotifier *hubUpdatedSubscription
)

func init() {
	hubUpdatedNotifier = &hubUpdatedSubscription{
		librariesUpdated: make(map[uint]database.Library),
	}
}

func onItemAddedCallback() {
	latest := make([]*models.LatestResult, 0, len(hubUpdatedNotifier.librariesUpdated))

	for _, library := range hubUpdatedNotifier.librariesUpdated {
		latestItems, err := database.GetLatestItemsFromLibrary(library, 24)
		if err != nil {
			log.Err(err).Msgf("Failed to get latest items from library %d", library.ID)

			return
		}

		var latestMediaItems []models.Item
		for _, latestItem := range latestItems {
			latestMediaItems = append(latestMediaItems, helpers.GetItemFromItemMetadata(latestItem))
		}

		latestResult := models.LatestResult{
			Library: &library,
			Items:   latestMediaItems,
		}

		latest = append(latest, &latestResult)
	}

	for _, observer := range utils.SubsciptionsManager.LatestItemsAddedObservers {
		observer <- latest
	}

	hubUpdatedNotifier.timer = nil
}

// OnLatestHubUpdated notifies all GraphQL clients that new items have been added to a library,
// prompting an update to the latest items.
func OnHubUpdated(library database.Library) {
	hubUpdatedNotifier.mu.Lock()

	if hubUpdatedNotifier.timer == nil {
		hubUpdatedNotifier.timer = time.AfterFunc(time.Second*5, onItemAddedCallback)
	}

	hubUpdatedNotifier.librariesUpdated[library.ID] = library

	hubUpdatedNotifier.mu.Unlock()
}
