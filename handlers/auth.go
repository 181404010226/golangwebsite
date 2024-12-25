package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	store  *sessions.CookieStore
	oauth  *oauth2.Config
	Client *mongo.Client
)

// InitStore initializes the session store
func InitStore() {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode, // 允许跨站点 cookie
	}
}

// InitOAuth 初始化 OAuth 配置
func InitOAuth() {
	oauth = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// SetClient 设置 MongoDB 客户端
func SetClient(client *mongo.Client) {
	Client = client
}

// GitHubCallbackHandler handles the GitHub OAuth callback
func GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := oauth.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user information from GitHub
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Username  string `json:"login"`
		AvatarURL string `json:"avatar_url"` // 添加这一行
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Store user information in the database
	collection := Client.Database("your-db-name").Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		map[string]interface{}{"github_id": user.ID},
		map[string]interface{}{
			"$set": map[string]interface{}{
				"name":       user.Name,
				"email":      user.Email,
				"username":   user.Username,
				"github_id":  user.ID,
				"avatar_url": user.AvatarURL, // 添加这一行
			},
		},
		options.Update().SetUpsert(true),
	)

	if err != nil {
		http.Error(w, "Failed to save user info", http.StatusInternalServerError)
		return
	}

	// Set session
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to React frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Default to local React dev server
	}
	log.Printf("用户登录成功 - ID: %d, 用户名: %s, 邮箱: %s", user.ID, user.Username, user.Email)
	http.Redirect(w, r, frontendURL+"?login_success=true", http.StatusFound)
}

// LoginHandler initiates the GitHub OAuth process
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// 检查用户是否已经登录
	session, _ := store.Get(r, "auth-session")
	// 在开头检查并输出登录状态
	if userID, ok := session.Values["user_id"].(int); ok {
		log.Printf("用户 ID %d 已登录", userID)
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		http.Redirect(w, r, frontendURL+"?login_success=true", http.StatusFound)
		return
	} else {
		log.Printf("用户未登录")
	}

	// 开始 OAuth 流程
	url := oauth.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// LogoutHandler clears the user session and redirects to the frontend
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	// 清除会话
	session.Options.MaxAge = -1 // 立即使会话过期
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// 获取前端的 URL 并重定向
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // 默认前端地址
	}
	http.Redirect(w, r, frontendURL, http.StatusFound)
}

// UserHandler returns the current user's information
func UserHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头，防止缓存
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Surrogate-Control", "no-store")

	session, err := store.Get(r, "auth-session")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		json.NewEncoder(w).Encode(map[string]interface{}{"user": nil})
		return
	}

	// 从数据库中获取用户信息
	collection := Client.Database("your-db-name").Collection("users")
	var user struct {
		ID        int    `json:"id" bson:"github_id"`
		Name      string `json:"name" bson:"name"`
		Email     string `json:"email" bson:"email"`
		Username  string `json:"username" bson:"username"`
		AvatarURL string `json:"avatar_url" bson:"avatar_url"`
	}

	err = collection.FindOne(context.Background(), map[string]interface{}{"github_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			json.NewEncoder(w).Encode(map[string]interface{}{"user": nil})
		} else {
			log.Printf("Failed to fetch user from database: %v", err)
			http.Error(w, "Failed to fetch user information", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"user": user})
}
