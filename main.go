package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"your-project/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 尝试加载 .env 文件，仅在本地环境中使用
	if err := godotenv.Load(); err != nil {
		log.Println("没有找到 .env 文件，继续使用系统环境变量")
	}

	// 连接到 MongoDB
	clientOptions := options.Client().ApplyURI(os.Getenv("DATABASE_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")

	// 将 MongoDB 客户端传递给处理器
	handlers.SetClient(client)

	// 设置 OAuth 重定向 URL
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	var host string
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		host = frontendURL
	} else {
		host = "http://localhost:" + port
	}

	redirectURL := host + "/auth/github/callback"
	os.Setenv("OAUTH_REDIRECT_URL", redirectURL)

	// 输出重定向 URL 以便调试
	log.Printf("OAuth Redirect URL: %s", redirectURL)

	// 初始化 session store
	handlers.InitStore()

	// 初始化 OAuth 配置
	handlers.InitOAuth()
	// Initialize session collection
	handlers.InitSessionCollection(client)
	// Initialize minutes collection
	handlers.InitMinutesCollection(client)

	// 设置路由
	r := mux.NewRouter()
	r.HandleFunc("/ws/sessions/{sessionId}", handlers.WebSocketHandler)
	r.HandleFunc("/api/user", handlers.UserHandler).Methods("GET")
	r.HandleFunc("/api/sessions", handlers.CreateSessionHandler).Methods("POST")
	r.HandleFunc("/api/sessions", handlers.GetSessionsHandler).Methods("GET")
	r.HandleFunc("/api/sessions/{sessionId}/start", handlers.StartMeetingHandler).Methods("POST")
	r.HandleFunc("/api/sessions/{sessionId}/comments", handlers.PostCommentHandler).Methods("POST")
	r.HandleFunc("/api/sessions/{sessionId}/comments", handlers.GetCommentsHandler).Methods("GET")
	r.HandleFunc("/api/sessions/{sessionId}", handlers.DeleteSessionHandler).Methods("DELETE")
	// Add new routes for meeting minutes
	r.HandleFunc("/api/sessions/{sessionId}/minutes", handlers.GetMinutesHandler).Methods("GET")
	r.HandleFunc("/api/sessions/{sessionId}/minutes", handlers.UpdateMinutesHandler).Methods("POST", "PUT")
	// r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("GET")
	r.HandleFunc("/auth/github/callback", handlers.GitHubCallbackHandler).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")

	// 静态文件服务
	staticPath := filepath.Join(".", "meeting-app", "build")
	indexPath := "index.html"
	spa := handlers.SpaHandler{
		StaticPath: staticPath,
		IndexPath:  indexPath,
	}
	r.PathPrefix("/").Handler(spa)

	// 创建一个新的 CORS 处理器
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8080",
			"http://localhost:3000",
			"ws://localhost:8080",
			"wss://localhost:8080", // 添加WSS协议
			os.Getenv("FRONTEND_URL"),
			strings.Replace(os.Getenv("FRONTEND_URL"), "http", "ws", 1),   // 添加对应的WS URL
			strings.Replace(os.Getenv("FRONTEND_URL"), "https", "wss", 1), // 添加对应的WSS URL
		},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})

	// 使用 CORS 中间件包装你的路由器
	handler := c.Handler(r)

	// 使用���的 handler 启动服务器
	log.Printf("Server is running on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
