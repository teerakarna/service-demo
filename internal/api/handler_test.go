package api_test

import (
	"bytes"
	"encoding/json"
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

func newRouter() *gin.Engine {
	return api.NewRouter(store.NewMemoryStore())
}

func TestHealthz(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReadyz(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateItem(t *testing.T) {
	r := newRouter()
	body := `{"name":"widget","description":"a test widget"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var item map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
	assert.Equal(t, "widget", item["name"])
	assert.NotEmpty(t, item["id"])
}

func TestCreateItemMissingName(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(`{"description":"no name"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetItem(t *testing.T) {
	r := newRouter()

	// Create first
	body := `{"name":"widget"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var item map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
	id := item["id"].(string)

	// Then get
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/items/"+id, nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestGetItemNotFound(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/items/nonexistent", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteItem(t *testing.T) {
	r := newRouter()

	body := `{"name":"to-delete"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var item map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
	id := item["id"].(string)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/items/"+id, nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNoContent, w2.Code)
}
