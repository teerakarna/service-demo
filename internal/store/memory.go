package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore is a thread-safe in-memory implementation of Store.
// Suitable for demos and tests; does not persist across restarts.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Item)}
}

func (s *MemoryStore) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		result = append(result, item)
	}
	return result
}

func (s *MemoryStore) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *MemoryStore) Create(name, description string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Item{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
	s.items[item.ID] = item
	return item
}

func (s *MemoryStore) Update(id, name, description string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}
	item.Name = name
	item.Description = description
	s.items[id] = item
	return item, true
}

func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[id]
	if ok {
		delete(s.items, id)
	}
	return ok
}
