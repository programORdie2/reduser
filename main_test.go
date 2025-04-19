package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func sendRequest(t *testing.T, handler http.HandlerFunc, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestSetVariableIntWithValidToken(t *testing.T) {
	payload := map[string]any{
		"token":          authToken,
		"action":         "set variable",
		"variable_name":  "counter",
		"variable_value": 42,
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}
	if response["status"] != "OK" {
		t.Fatalf("expected OK response, got %s", response["status"])
	}
}

func TestGetVariableIntWithValidToken(t *testing.T) {
	// Set first
	variables["counter"] = 42

	payload := map[string]any{
		"token":         authToken,
		"action":        "get variable",
		"variable_name": "counter",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}

	if response["type"] != "int" {
		t.Fatalf("expected int, got %T", response["value"])
	}

	val, ok := response["value"].(int64)
	if !ok {
		// If it's not an int64, it must be a float64
		new_val, ok := response["value"].(float64)
		if !ok {
			t.Fatalf("expected int or float64, got %T", response["value"])
		}

		val = int64(new_val)
	}

	if val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}
}

func TestSetVariableFloatWithValidToken(t *testing.T) {
	payload := map[string]any{
		"token":          authToken,
		"action":         "set variable",
		"variable_name":  "counter",
		"variable_value": 42.5,
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}
	if response["status"] != "OK" {
		t.Fatalf("expected OK response, got %s", response["status"])
	}
}

func TestGetVariableFloatWithValidToken(t *testing.T) {
	// Set first
	variables["counter"] = 42.5

	payload := map[string]any{
		"token":         authToken,
		"action":        "get variable",
		"variable_name": "counter",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}

	if response["type"] != "float64" {
		t.Fatalf("expected float64, got %T", response["value"])
	}

	val, ok := response["value"].(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", response["value"])
	}

	if val != 42.5 {
		t.Fatalf("expected 42.5, got %f", val)
	}
}

func TestSetVariableStringWithValidToken(t *testing.T) {
	payload := map[string]any{
		"token":          authToken,
		"action":         "set variable",
		"variable_name":  "hello",
		"variable_value": "world",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}
	if response["status"] != "OK" {
		t.Fatalf("expected OK response, got %s", response["status"])
	}
}

func TestGetVariableStringWithValidToken(t *testing.T) {
	// Set first
	variables["hello"] = "world"

	payload := map[string]any{
		"token":         authToken,
		"action":        "get variable",
		"variable_name": "hello",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.Code)
	}

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("expected OK response, got %s", res.Body.String())
	}

	if response["type"] != "string" {
		t.Fatalf("expected string, got %T", response["value"])
	}

	val, ok := response["value"].(string)
	if !ok {
		t.Fatalf("expected string, got %T", response["value"])
	}

	if val != "world" {
		t.Fatalf("expected 'world', got %s", val)
	}
}

func TestSetVariableWithInvalidToken(t *testing.T) {
	payload := map[string]any{
		"token":          "wrongtoken",
		"action":         "set variable",
		"variable_name":  "counter",
		"variable_value": 100,
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", res.Code)
	}
}

func TestGetVariableWithInvalidToken(t *testing.T) {
	payload := map[string]any{
		"token":         "invalid",
		"action":        "get variable",
		"variable_name": "counter",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", res.Code)
	}
}

func TestGetUnknownVariable(t *testing.T) {
	payload := map[string]any{
		"token":         authToken,
		"action":        "get variable",
		"variable_name": "nonexistent",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", res.Code)
	}
}

func TestUnknownAction(t *testing.T) {
	payload := map[string]any{
		"token":         authToken,
		"action":        "delete variable",
		"variable_name": "counter",
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unknown action, got %d", res.Code)
	}
}

func TestMissingVariableNameOnSet(t *testing.T) {
	payload := map[string]any{
		"token":          authToken,
		"action":         "set variable",
		"variable_value": 100,
	}

	res := sendRequest(t, handleRequest, payload)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", res.Code)
	}
}

func TestInvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handleRequest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	handleRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for invalid JSON, got %d", w.Code)
	}
}
