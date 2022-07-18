package utils

import (
	"sync"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/models"
)

type Subscriptions struct {
	ItemAddedObservers        map[string]chan models.Item
	ItemUpdatedObservers      map[string]chan models.Item
	LibraryAddedObservers     map[string]chan *database.Library
	LatestItemsAddedObservers map[string]chan []*models.LatestResult
	mu                        sync.Mutex
}

func (s *Subscriptions) Lock() {
	s.mu.Lock()
}

func (s *Subscriptions) Unlock() {
	s.mu.Unlock()
}

var SubsciptionsManager *Subscriptions

func init() {
	SubsciptionsManager = &Subscriptions{
		ItemAddedObservers:        make(map[string]chan models.Item),
		ItemUpdatedObservers:      make(map[string]chan models.Item),
		LibraryAddedObservers:     make(map[string]chan *database.Library),
		LatestItemsAddedObservers: make(map[string]chan []*models.LatestResult),
	}
}
