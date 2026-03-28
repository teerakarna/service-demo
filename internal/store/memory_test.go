package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teerakarna/service-demo/internal/store"
)

func TestCreateAndGet(t *testing.T) {
	s := store.NewMemoryStore()
	item := s.Create("widget", "a test widget")
	require.NotEmpty(t, item.ID)
	assert.Equal(t, "widget", item.Name)

	got, ok := s.Get(item.ID)
	require.True(t, ok)
	assert.Equal(t, item.ID, got.ID)
}

func TestGetNotFound(t *testing.T) {
	s := store.NewMemoryStore()
	_, ok := s.Get("nonexistent")
	assert.False(t, ok)
}

func TestUpdate(t *testing.T) {
	s := store.NewMemoryStore()
	item := s.Create("original", "")
	updated, ok := s.Update(item.ID, "updated", "new desc")
	require.True(t, ok)
	assert.Equal(t, "updated", updated.Name)
}

func TestUpdateNotFound(t *testing.T) {
	s := store.NewMemoryStore()
	_, ok := s.Update("nonexistent", "x", "y")
	assert.False(t, ok)
}

func TestDelete(t *testing.T) {
	s := store.NewMemoryStore()
	item := s.Create("to-delete", "")
	assert.True(t, s.Delete(item.ID))
	assert.False(t, s.Delete(item.ID)) // second delete returns false
}

func TestList(t *testing.T) {
	s := store.NewMemoryStore()
	s.Create("a", "")
	s.Create("b", "")
	assert.Len(t, s.List(), 2)
}
