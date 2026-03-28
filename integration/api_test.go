// Integration tests spin up a real httptest.Server and exercise the full
// HTTP stack: routing, middleware, and storage together.
// Run with: go test ./integration/...
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teerakarna/service-demo/internal/api"
	"github.com/teerakarna/service-demo/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newServer() *httptest.Server {
	return httptest.NewServer(api.NewRouter(store.NewMemoryStore()))
}

func TestHealthEndpoints(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	for _, path := range []string{"/healthz", "/readyz"} {
		resp, err := http.Get(srv.URL + path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, path)
	}
}

// TestItemLifecycle exercises the full CRUD flow in sequence.
func TestItemLifecycle(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	client := srv.Client()

	// Create
	body := `{"name":"widget","description":"integration test item"}`
	resp, err := client.Post(srv.URL+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var item map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	resp.Body.Close()
	id := item["id"].(string)

	// Get
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/items/%s", srv.URL, id))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// List
	resp, err = client.Get(srv.URL + "/api/v1/items")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Len(t, list, 1)
	resp.Body.Close()

	// Update
	update := `{"name":"updated-widget","description":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/items/%s", srv.URL, id), bytes.NewBufferString(update))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Delete
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/items/%s", srv.URL, id), nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify gone
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/items/%s", srv.URL, id))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateValidation(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	// Missing required field 'name'
	resp, err := srv.Client().Post(srv.URL+"/api/v1/items", "application/json", bytes.NewBufferString(`{"description":"no name"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
