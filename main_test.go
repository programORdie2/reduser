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

	projectToken := ""

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

		// Parse response
		var project struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Token string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&project)
		assert.Equal(t, 0, project.ID)
		assert.Equal(t, "MyProject", project.Name)
		assert.NotEmpty(t, project.Token)
		projectToken = project.Token
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

	// Set variable
	t.Run("Set variable", func(t *testing.T) {
		body := `{"variable":"var1","value":"hi","table":1,"token":"` + projectToken + `","action":"set"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/access", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Get variable
	t.Run("Get variable", func(t *testing.T) {
		body := `{"variable":"var1","table":1,"token":"` + projectToken + `","action":"get"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/access", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var respBody struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, "hi", respBody.Value)
		assert.Equal(t, "string", respBody.Type)
	})

	// Update variable
	t.Run("Update variable", func(t *testing.T) {
		body := `{"new_type":"int"}`
		req, _ := http.NewRequest("PUT", server.URL+"/api/projects/1/tables/1/variables/var1", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Set variable int
	t.Run("Set variable int", func(t *testing.T) {
		body := `{"variable":"var1","value":123,"table":1,"token":"` + projectToken + `","action":"set"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/access", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Get variable int
	t.Run("Get variable int", func(t *testing.T) {
		body := `{"variable":"var1","table":1,"token":"` + projectToken + `","action":"get"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/access", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var respBody struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		}
		json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, 123, respBody.Value)
		assert.Equal(t, "int", respBody.Type)
	})

	// Set variable with wrong type
	t.Run("Set variable with wrong type", func(t *testing.T) {
		body := `{"variable":"var1","value":"hi","table":1,"token":"` + projectToken + `","action":"set"}`
		req, _ := http.NewRequest("POST", server.URL+"/api/access", bytes.NewBufferString(body))
		resp, _ := http.DefaultClient.Do(req)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
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
