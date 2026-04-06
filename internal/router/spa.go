package router

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
)

func newSPAHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := fsys.Open(path)
		if err != nil {
			serveIndexHTML(fsys, w)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}

func newServiceWorkerHandler(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Service-Worker-Allowed", "/")
		http.FileServerFS(fsys).ServeHTTP(w, r)
	}
}

func serveIndexHTML(fsys fs.FS, w http.ResponseWriter) {
	indexFile, err := fsys.Open("index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	defer indexFile.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, indexFile) //nolint:errcheck
}
