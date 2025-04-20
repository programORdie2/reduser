package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthAndProjectFlow(t *testing.T) {
	// Setup in memory DB
	db := InitDB(":memory:?cache=shared")
	defer db.Close()

	// Set up test server
	router := chi.NewRouter() // this should return your *http.ServeMux or chi.Router
	MountAPIRoutes(router, db)
	server := httptest.NewServer(router)
	defer server.Close()

	// Register
	resp, err := http.Post(server.URL+"/api/register", "application/json", bytes.NewBufferString(`{"username":"testuser","password":"1234"}`))
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Login
	var loginResp struct {
		Token string `json:"token"`
	}
	resp, err = http.Post(server.URL+"/api/login", "application/json", bytes.NewBufferString(`{"username":"testuser","password":"1234"}`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&loginResp)
	assert.NotEmpty(t, loginResp.Token)
	token := loginResp.Token

	// Fake JWT Test
	t.Run("Rejects fake JWTs", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/projects", nil)
		req.Header.Set("Authorization", "Bearer faketokenlol")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Load projects (empty initially)
	t.Run("List projects", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL+"/api/projects", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var projects []interface{}
		json.NewDecoder(resp.Body).Decode(&projects)
		assert.Len(t, projects, 0)
	})

	// Create project
	t.Run("Create project", func(t *testing.T) {
		body := `{"name":"MyProject"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/projects", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Load project details
	t.Run("Load project", func(t *testing.T) {
		// Assuming ID = 1 for simplicity
		req, _ := http.NewRequest("GET", server.URL+"/api/projects/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Create table
	t.Run("Create table", func(t *testing.T) {
		body := `{"name":"MyTable"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/projects/1/tables", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Add variable
	t.Run("Add variable", func(t *testing.T) {
		body := `{"name":"var1","type":"string","value":""}`
		req, _ := http.NewRequest("POST", server.URL+"/api/projects/1/tables/1/variables", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Update variable
	t.Run("Update variable", func(t *testing.T) {
		body := `{"new_name":"renamedVar","new_type":"int"}`
		req, _ := http.NewRequest("PUT", server.URL+"/api/projects/1/tables/1/variables/var1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Delete variable
	t.Run("Delete variable", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/projects/1/tables/1/variables/renamedVar", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Rename table
	t.Run("Update table name", func(t *testing.T) {
		body := `{"name":"RenamedTable"}`
		req, _ := http.NewRequest("PUT", server.URL+"/api/projects/1/tables/1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Rename project
	t.Run("Update project name", func(t *testing.T) {
		body := `{"name":"RenamedProject"}`
		req, _ := http.NewRequest("PUT", server.URL+"/api/projects/1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Delete table
	t.Run("Delete table", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/projects/1/tables/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Delete project
	t.Run("Delete project", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", server.URL+"/api/projects/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
