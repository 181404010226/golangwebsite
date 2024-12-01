// your-project/handlers/comment.go
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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitCommentCollection(client *mongo.Client) {
    // Initialize comment collection if separate, else use sessionCollection
}

// PostCommentHandler handles posting a new comment
func PostCommentHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    sessionID := vars["sessionId"]

    // 获取用户会话
    session, err := store.Get(r, "auth-session")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    userID, ok := session.Values["user_id"].(int)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // 转换 sessionID 为 ObjectID
    objectID, err := primitive.ObjectIDFromHex(sessionID)
    if err != nil {
        http.Error(w, "Invalid session ID format", http.StatusBadRequest)
        return
    }

    var commentInput struct {
        Content string `json:"content"`
        Stars   int    `json:"stars"`
    }

    if err := json.NewDecoder(r.Body).Decode(&commentInput); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // 验证星级
    if commentInput.Stars < 1 || commentInput.Stars > 10 {
        http.Error(w, "Stars must be between 1 and 10", http.StatusBadRequest)
        return
    }

    // 获取用户信息用于广播
    var user struct {
        ID        primitive.ObjectID `bson:"_id"`
        Username  string            `bson:"username"`
        AvatarURL string            `bson:"avatar_url"`
    }
    err = Client.Database("your-db-name").Collection("users").FindOne(
        context.Background(),
        bson.M{"github_id": userID},
    ).Decode(&user)

    if err != nil {
        log.Printf("Error fetching user info: %v", err)
        http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
        return
    }

    comment := models.Comment{
        ID:        primitive.NewObjectID(),
        UserID:    user.ID,           // 使用用户在 MongoDB 中的 _id
        Username:  user.Username,
        AvatarURL: user.AvatarURL,
        Content:   commentInput.Content,
        Stars:     commentInput.Stars,
        CreatedAt: time.Now(),
    }

    // 更新数据库
    filter := bson.M{"_id": objectID}

    // 使用 $push 将新评论添加到最后一个 summary 的 comments 数组中
    update := bson.M{
        "$push": bson.M{
            "summaries.$[].comments": comment,
        },
    }

    opts := options.Update().SetUpsert(false)
    _, err = sessionCollection.UpdateOne(
        context.Background(),
        filter,
        update,
        opts,
    )

    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Failed to add comment", http.StatusInternalServerError)
        return
    }

    // 广播带有用户信息的评论
    broadcastComment := map[string]interface{}{
        "type": "newComment",
        "comment": map[string]interface{}{
            "id":         comment.ID.Hex(),
            "content":    comment.Content,
            "stars":      comment.Stars,
            "created_at": comment.CreatedAt,
            "username":   comment.Username,     // 直接放在顶层
            "avatar_url": comment.AvatarURL,    // 直接放在顶层
            "user_id":    comment.UserID.Hex(), // 如果需要的话
        },
    }

    // 使用单独的 goroutine 进行广播
    go func() {
        Broadcast(sessionID, broadcastComment)
    }()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "comment": comment,
    })
}

// GetCommentsHandler retrieves comments for a session
func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    sessionID := vars["sessionId"]

    // 转换 sessionID 为 ObjectID
    objectID, err := primitive.ObjectIDFromHex(sessionID)
    if err != nil {
        http.Error(w, "Invalid session ID format", http.StatusBadRequest)
        return
    }

    var session models.Session
    err = sessionCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&session)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }

    // 构建所有评论的列表
    allComments := []models.Comment{}
    for _, summary := range session.Summaries {
        allComments = append(allComments, summary.Comments...)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(allComments)
}