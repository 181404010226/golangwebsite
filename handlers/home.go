package handlers

import (
    "html/template"
    "net/http"
    //"os"

    //"github.com/gorilla/sessions"
)

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

// HomeHandler handles requests to the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    session, err := store.Get(r, "auth-session")
    if err != nil {
        http.Error(w, "Failed to get session", http.StatusInternalServerError)
        return
    }

    user, _ := session.Values["user"].(string)
    avatarURL, _ := session.Values["avatar_url"].(string) // 添加这一行

    data := map[string]interface{}{
        "User": user,
        "AvatarURL": avatarURL, // 添加这一行
    }

    if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}