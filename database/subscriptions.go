package database

import (
	"sync"
)

type Subscriptions struct {
	ItemAddedObservers    map[string]chan *ItemMetadata
	ItemUpdatedObservers  map[string]chan *ItemMetadata
	LibraryAddedObservers map[string]chan *Library
	mu                    sync.Mutex
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
		ItemAddedObservers:    make(map[string]chan *ItemMetadata),
		ItemUpdatedObservers:  make(map[string]chan *ItemMetadata),
		LibraryAddedObservers: make(map[string]chan *Library),
	}
}

func SendItemUpdated(item ItemMetadata) {
	for _, observer := range SubsciptionsManager.ItemUpdatedObservers {
		observer <- &item
	}
}
