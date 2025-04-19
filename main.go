package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

const (
	authToken = "stiostudio"
)

type request struct {
	Token         string `json:"token"`
	Action        string `json:"action"`
	VariableName  string `json:"variable_name"`
	VariableValue any    `json:"variable_value,omitempty"`
}

var (
	variables = make(map[string]any)
	mu        sync.RWMutex
)

func main() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "{\"error\": \"Method not allowed\"}", http.StatusMethodNotAllowed)
		return
	}

	var req request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "{\"error\": \"Invalid JSON\"}", http.StatusBadRequest)
		return
	}

	if req.Token != authToken {
		http.Error(w, "{\"error\": \"Unauthorized\"}", http.StatusUnauthorized)
		return
	}

	switch req.Action {
	case "get variable":
		mu.RLock()
		val, ok := variables[req.VariableName]
		mu.RUnlock()

		if !ok {
			http.Error(w, "{\"error\": \"Variable not found\"}", http.StatusNotFound)
			return
		}

		resp := map[string]any{
			"value": val,
			"type":  fmt.Sprintf("%T", val),
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)

	case "set variable":
		if req.VariableName == "" {
			http.Error(w, "{\"error\": \"Missing variable_name\"}", http.StatusBadRequest)
			return
		}

		mu.Lock()
		variables[req.VariableName] = req.VariableValue
		mu.Unlock()

		resp := map[string]string{
			"status": "OK",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}
