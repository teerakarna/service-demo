package store

import "time"

// Item is the core domain object.
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Store is the storage interface. Swap the implementation without touching handlers.
type Store interface {
	List() []Item
	Get(id string) (Item, bool)
	Create(name, description string) Item
	Update(id, name, description string) (Item, bool)
	Delete(id string) bool
}
