package database

import (
	"sync"

	"github.com/meteorae/meteorae-server/sdk"
)

type Subscriptions struct {
	ItemAddedObservers      map[string]chan sdk.Item
	ItemUpdatedObservers    map[string]chan sdk.Item
	LibraryAddedObservers   map[string]chan *Library
	LibraryUpdatedObservers map[string]chan *Library
	mu                      sync.Mutex
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
		ItemAddedObservers:      make(map[string]chan sdk.Item),
		ItemUpdatedObservers:    make(map[string]chan sdk.Item),
		LibraryAddedObservers:   make(map[string]chan *Library),
		LibraryUpdatedObservers: make(map[string]chan *Library),
	}
}
