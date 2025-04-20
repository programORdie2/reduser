package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	db := InitDB("app.db")
	defer db.Close()

	r := chi.NewRouter()
	MountAPIRoutes(r, db)

	log.Println("listening on localhost:8080")
	http.ListenAndServe("localhost:8080", r)
}
