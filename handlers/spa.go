package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

type SpaHandler struct {
	StaticPath string
	IndexPath  string
}

func (h SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取请求的文件路径
	path := filepath.Join(h.StaticPath, r.URL.Path)

	// 检查文件是否存在
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// 如果文件不存在，返回 index.html
		http.ServeFile(w, r, filepath.Join(h.StaticPath, h.IndexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 提供静态文件服务
	http.FileServer(http.Dir(h.StaticPath)).ServeHTTP(w, r)
}
