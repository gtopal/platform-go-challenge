package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

func main() {
	// Add a default user for demo/testing
	defaultID := uuid.New()
	log.Printf("Default user_id: %s\n", defaultID)
	store.AddUser(&User{ID: defaultID})
	setupRoutes()
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// TODO: ADD PAGINATION TO GET FAVOURITES
// TODO: TRY TO CREATE ONE ENDPOINT FOR ADD/REMOVE/EDIT USING HTTP VERBS
