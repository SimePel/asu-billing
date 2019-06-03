package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	t := template.Must(template.New("user").ParseFiles("templates/usr/index.html"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "index", nil)
	})

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "assets")
	fileServer(r, "/assets", http.Dir(filesDir))

	http.ListenAndServe(":8080", r)
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
