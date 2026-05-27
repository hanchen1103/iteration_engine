package memory

import (
	"sync"

	"github.com/hanchen1103/iteration_engine/domain"
	"github.com/hanchen1103/iteration_engine/ports"
)

type Store struct {
	mu            sync.RWMutex
	nextRunID     int64
	nextVersionID int64
	nextEventID   int64
	runs          map[string]*domain.Run
	versions      map[string]*domain.Version
	events        map[string][]*domain.Event
}

var _ ports.Store = (*Store)(nil)

func NewStore() *Store {
	return &Store{
		runs:     map[string]*domain.Run{},
		versions: map[string]*domain.Version{},
		events:   map[string][]*domain.Event{},
	}
}
