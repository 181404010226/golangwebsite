// your-project/handlers/session.go
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"your-project/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var sessionCollection *mongo.Collection

func InitSessionCollection(client *mongo.Client) {
	sessionCollection = client.Database("your-db-name").Collection("sessions")
}

// CreateSessionHandler creates a new session
func CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	var session models.Session
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// 初始化基本字段
	session.ID = primitive.NewObjectID()
	session.CreatedAt = time.Now()

	// 创建一个初始的 summary
	initialSummary := models.Summary{
		ParticipantID: primitive.NilObjectID, // 或者从请求中获取参与者ID
		Content:       "",                    // 初始内容为空
		Comments:      []models.Comment{},    // 显式初始化空的评论数组
		CreatedAt:     time.Now(),            // 设置创建时间
	}

	// 初始化 summaries 数组并添加初始 summary
	session.Summaries = []models.Summary{initialSummary}

	// 插入到数据库
	result, err := sessionCollection.InsertOne(context.Background(), session)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": result.InsertedID,
	})
}

// GetSessionsHandler retrieves all sessions
func GetSessionsHandler(w http.ResponseWriter, r *http.Request) {
	cursor, err := sessionCollection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var sessions []models.Session
	if err = cursor.All(context.Background(), &sessions); err != nil {
		http.Error(w, "Failed to parse sessions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(sessions)
}

// DeleteSessionHandler deletes a session by its ID
func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

	// Convert sessionID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}
	result, err := sessionCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Session deleted successfully"})
}
