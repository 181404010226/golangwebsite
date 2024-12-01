// your-project/handlers/websocket.go
package handlers

import (
	"log"
	"net/http"
	"time"

	"context"
	"os"
	"sort"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type MeetingClient struct {
	conn      *websocket.Conn
	sessionID string
	userID    int
	username  string
	avatarURL string
	joinedAt  time.Time
}

var MeetingClients = make(map[string][]*MeetingClient) // sessionID -> MeetingClients

func Broadcast(sessionID string, message interface{}) {
	for _, client := range MeetingClients[sessionID] {
		err := client.conn.WriteJSON(message)
		if err != nil {
			log.Println("WebSocket write error:", err)
		}
	}
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]

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

	// 检查用户是否已经在会话中
	clients := MeetingClients[sessionID]
	for _, existingClient := range clients {
		if existingClient.userID == userID {
			// 关闭现有连接
			existingClient.conn.Close()
			removeClient(sessionID, existingClient)
			break
		}
	}

	// 获取用户信息
	var user struct {
		Username  string `bson:"username"`
		AvatarURL string `bson:"avatar_url"`
	}

	err = Client.Database("your-db-name").Collection("users").FindOne(
		context.Background(),
		bson.M{"github_id": userID},
	).Decode(&user)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 设置 CORS headers
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// 获取请求的Origin
		origin := r.Header.Get("Origin")

		// 允许的域名列表
		allowedOrigins := []string{
			"http://localhost:8080",
			"http://localhost:3000",
			"ws://localhost:8080",
			"wss://localhost:8080",
			os.Getenv("FRONTEND_URL"),
		}

		// 检查Origin是否在允许列表中
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		log.Printf("Rejected WebSocket connection from origin: %s", origin)
		return false
	}

	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	client := &MeetingClient{
		conn:      conn,
		sessionID: sessionID,
		userID:    userID,
		username:  user.Username,
		avatarURL: user.AvatarURL,
		joinedAt:  time.Now(),
	}

	// 添加新客户端到会话
	if MeetingClients[sessionID] == nil {
		MeetingClients[sessionID] = make([]*MeetingClient, 0)
	}
	MeetingClients[sessionID] = append(MeetingClients[sessionID], client)

	// 广播更新后的参与者列表
	broadcastParticipantsList(sessionID)

	defer func() {
		conn.Close()
		removeClient(sessionID, client)
		broadcastParticipantsList(sessionID)
	}()

	// 发送连接成功消息
	if err := conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "Successfully connected to session",
	}); err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// 设置更合理的超时时间
	conn.SetReadLimit(1024 * 1024) // 1MB
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 处理消息循环
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			}
			break
		}

		// 处理 ping 消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg["type"] == "ping" {
			// 发送 pong 响应
			pongMsg := map[string]string{"type": "pong"}
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Error sending pong: %v", err)
				break
			}
			continue
		}

		// 处理其他消息类型
		handleWebSocketMessage(sessionID, msg)
	}
}

func removeClient(sessionID string, client *MeetingClient) {
	clients := MeetingClients[sessionID]
	for i, c := range clients {
		if c == client {
			MeetingClients[sessionID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	// 如果没有客户端了，清理会话
	if len(MeetingClients[sessionID]) == 0 {
		delete(MeetingClients, sessionID)
	}
}

func handleWebSocketMessage(sessionID string, msg map[string]interface{}) {
	switch msg["type"] {
	case "joinSession":
		log.Printf("Client joined session: %s", sessionID)
		// 立即广播更新后的参与者列表
		broadcastParticipantsList(sessionID)
	case "summarySubmitted":
		log.Printf("Summary submitted in session: %s", sessionID)
		Broadcast(sessionID, msg)
	case "newComment":
		log.Printf("New comment in session: %s", sessionID)
		Broadcast(sessionID, msg)
	default:
		log.Printf("Unknown message type received: %v", msg["type"])
	}
}

func broadcastParticipantsList(sessionID string) {
	// 使用 map 来确保每个用户只出现一次
	uniqueParticipants := make(map[int]map[string]interface{})

	for _, client := range MeetingClients[sessionID] {
		uniqueParticipants[client.userID] = map[string]interface{}{
			"id":        client.userID,
			"username":  client.username,
			"avatarUrl": client.avatarURL,
			"joinedAt":  client.joinedAt,
		}
	}

	// 将 map 转换为 slice
	participants := make([]map[string]interface{}, 0, len(uniqueParticipants))
	for _, participant := range uniqueParticipants {
		participants = append(participants, participant)
	}

	// 按加入时间排序
	sort.Slice(participants, func(i, j int) bool {
		iTime := participants[i]["joinedAt"].(time.Time)
		jTime := participants[j]["joinedAt"].(time.Time)
		return iTime.Before(jTime)
	})

	message := map[string]interface{}{
		"type":         "participantsList",
		"participants": participants,
	}

	Broadcast(sessionID, message)
}
