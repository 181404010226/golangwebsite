// your-project/handlers/meeting.go
package handlers

import (
	"context"
	"net/http"
	"your-project/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func NextParticipant(sessionID string) (*models.Participant, error) {
	// Fetch the session
	var session models.Session
	err := sessionCollection.FindOne(context.Background(), bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}

	for _, participant := range session.Participants {
		if !participant.Summarized {
			return &participant, nil
		}
	}

	return nil, nil // All participants have summarized
}

func StartMeetingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	participant, err := NextParticipant(sessionID)
	if err != nil {
		http.Error(w, "Failed to determine next participant", http.StatusInternalServerError)
		return
	}

	if participant != nil {
		// Notify clients via WebSocket
		Broadcast(sessionID, map[string]interface{}{
			"type":        "nextParticipant",
			"participant": participant,
		})
	} else {
		Broadcast(sessionID, map[string]interface{}{
			"type": "meetingEnded",
		})
	}

	w.WriteHeader(http.StatusOK)
}
