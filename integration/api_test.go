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
	"os"
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

// target returns the base URL to test against and a client to use.
// If BASE_URL is set, tests run against that already-running server (e.g. a
// deployed preprod pod, reached via port-forward). Otherwise they spin up an
// in-process httptest.Server, closed automatically when the test ends.
func target(t *testing.T) (string, *http.Client) {
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		return baseURL, http.DefaultClient
	}
	srv := httptest.NewServer(api.NewRouter(store.NewMemoryStore()))
	t.Cleanup(srv.Close)
	return srv.URL, srv.Client()
}

func TestHealthEndpoints(t *testing.T) {
	baseURL, client := target(t)

	for _, path := range []string{"/healthz", "/readyz"} {
		resp, err := client.Get(baseURL + path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode, path)
	}
}

// TestItemLifecycle exercises the full CRUD flow in sequence.
func TestItemLifecycle(t *testing.T) {
	baseURL, client := target(t)

	// Create
	body := `{"name":"widget","description":"integration test item"}`
	resp, err := client.Post(baseURL+"/api/v1/items", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var item map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&item))
	resp.Body.Close()
	id := item["id"].(string)

	// Get
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/items/%s", baseURL, id))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// List
	resp, err = client.Get(baseURL + "/api/v1/items")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list []interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	resp.Body.Close()

	// Update
	update := `{"name":"updated-widget","description":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/items/%s", baseURL, id), bytes.NewBufferString(update))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Delete
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/items/%s", baseURL, id), nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// Verify gone
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/items/%s", baseURL, id))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateValidation(t *testing.T) {
	baseURL, client := target(t)

	// Missing required field 'name'
	resp, err := client.Post(baseURL+"/api/v1/items", "application/json", bytes.NewBufferString(`{"description":"no name"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
