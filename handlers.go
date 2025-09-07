package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// parseUserID extracts and parses the user_id query parameter as uuid.UUID
func parseUserID(r *http.Request, w http.ResponseWriter) (uuid.UUID, bool) {
	userIDStr := r.URL.Query().Get("user_id")
	log.Printf("Parsed user_id: %s of type %T\n", userIDStr, userIDStr)
	userID, err := uuid.Parse(userIDStr)
	log.Printf("Parsed userID: %s of type %T\n", userID, userID)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return uuid.UUID{}, false
	}
	return userID, true
}

// parseAssetID extracts and parses the asset_id query parameter as uuid.UUID
func parseAssetID(r *http.Request, w http.ResponseWriter) (uuid.UUID, bool) {
	assetIDStr := r.URL.Query().Get("asset_id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		http.Error(w, "Invalid asset_id", http.StatusBadRequest)
		return uuid.UUID{}, false
	}
	return assetID, true
}

func setupRoutes() {
	http.HandleFunc("/favourites", handleFavourites)
	http.HandleFunc("/favourites/add", handleAddFavourite)
	http.HandleFunc("/favourites/remove", handleRemoveFavourite)
	http.HandleFunc("/favourites/edit", handleEditFavourite)
	http.HandleFunc("/favourites/delete", handleDeleteFavourite)

}

// List all the assets of the user with Favorite == true
func handleFavourites(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(r, w)
	if !ok {
		return
	}
	user := store.GetUser(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Only return assets with Favorite == true
	favs := make([]Asset, 0)
	for _, asset := range user.Favourites {
		if asset.IsFavorite() {
			favs = append(favs, asset)
		}
	}
	json.NewEncoder(w).Encode(favs)
}

// Create an asset
func handleAddFavourite(w http.ResponseWriter, r *http.Request) {

	userID, ok := parseUserID(r, w)
	if !ok {
		return
	}
	user := store.GetUser(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	var req struct {
		Type     string          `json:"type"`
		Asset    json.RawMessage `json:"asset"`
		Favorite bool            `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	var asset Asset
	switch req.Type {
	case ChartType:
		log.Printf("Adding Chart asset: %s\n", req.Asset)
		var c Chart
		if err := json.Unmarshal(req.Asset, &c); err != nil {
			http.Error(w, "Invalid chart asset", http.StatusBadRequest)
			return
		}
		if c.ID == uuid.Nil {
			c.ID = uuid.New()
		}
		c.Favorite = req.Favorite
		asset = &c
	case InsightType:
		log.Printf("Adding Insight asset: %s\n", req.Asset)
		var i Insight
		if err := json.Unmarshal(req.Asset, &i); err != nil {
			http.Error(w, "Invalid insight asset", http.StatusBadRequest)
			return
		}
		if i.ID == uuid.Nil {
			i.ID = uuid.New()
		}
		i.Favorite = req.Favorite
		asset = &i
	case AudienceType:
		log.Printf("Adding Audience asset: %s\n", req.Asset)
		var a Audience
		if err := json.Unmarshal(req.Asset, &a); err != nil {
			http.Error(w, "Invalid audience asset", http.StatusBadRequest)
			return
		}
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		a.Favorite = req.Favorite
		asset = &a
	default:
		http.Error(w, "Unknown asset type", http.StatusBadRequest)
		return
	}
	user.Favourites = append(user.Favourites, asset)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(asset)
}

// Edit the isFavorite field of an asset
func handleRemoveFavourite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := parseUserID(r, w)
	if !ok {
		return
	}
	assetID, ok := parseAssetID(r, w)
	if !ok {
		return
	}
	user := store.GetUser(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	var req struct {
		Favorite bool `json:"favorite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	for _, fav := range user.Favourites {
		if fav.GetID() == assetID {
			fav.SetFavorite(req.Favorite)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(fav)
			return
		}
	}
	http.Error(w, "Asset not found in favourites", http.StatusNotFound)
}

// Edits the description of an asset in general
func handleEditFavourite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := parseUserID(r, w)
	if !ok {
		return
	}
	assetID, ok := parseAssetID(r, w)
	if !ok {
		return
	}
	user := store.GetUser(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	var req struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	for _, fav := range user.Favourites {
		if fav.GetID() == assetID {
			fav.SetDescription(req.Description)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(fav)
			return
		}
	}
	http.Error(w, "Asset not found in favourites", http.StatusNotFound)
}

// Deletes an asset in general
func handleDeleteFavourite(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(r, w)
	if !ok {
		return
	}
	assetID, ok := parseAssetID(r, w)
	if !ok {
		return
	}
	user := store.GetUser(userID)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	found := false
	newFavs := make([]Asset, 0, len(user.Favourites))
	for _, fav := range user.Favourites {
		if fav.GetID() == assetID {
			found = true
			continue
		}
		newFavs = append(newFavs, fav)
	}
	if !found {
		http.Error(w, "Asset not found in favourites", http.StatusNotFound)
		return
	}
	user.Favourites = newFavs
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Favourites)
}
