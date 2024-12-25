package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"your-project/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var minutesCollection *mongo.Collection

func InitMinutesCollection(client *mongo.Client) {
	minutesCollection = client.Database("your-db-name").Collection("minutes")
}

// GetMinutesHandler retrieves meeting minutes for a session
func GetMinutesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Convert sessionID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Find minutes for the session
	var minutes models.Minutes
	err = minutesCollection.FindOne(context.Background(), bson.M{"session_id": objectID}).Decode(&minutes)
	if err == mongo.ErrNoDocuments {
		// If no minutes exist, return an empty minutes object
		minutes = models.Minutes{
			SessionID: objectID,
			Content:   "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	} else if err != nil {
		http.Error(w, "Failed to fetch minutes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(minutes)
}

// UpdateMinutesHandler updates meeting minutes for a session
func UpdateMinutesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Get user from session
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Convert IDs
	sessionObjectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	userObjectID := primitive.NewObjectID() // You might need to convert the userID to ObjectID differently

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"content":    input.Content,
			"updated_at": now,
			"updated_by": userObjectID,
		},
	}

	// Try to update existing minutes
	result, err := minutesCollection.UpdateOne(
		context.Background(),
		bson.M{"session_id": sessionObjectID},
		update,
	)

	// If no document was updated, create new minutes
	if err != nil || result.MatchedCount == 0 {
		minutes := models.Minutes{
			ID:        primitive.NewObjectID(),
			SessionID: sessionObjectID,
			Content:   input.Content,
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: userObjectID,
			UpdatedBy: userObjectID,
		}

		_, err = minutesCollection.InsertOne(context.Background(), minutes)
		if err != nil {
			http.Error(w, "Failed to create minutes", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
