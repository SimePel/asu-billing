package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("./templates/usr/")))
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(jsonContentType)
			r.Use(userCtx)
			r.Get("/", getUser)
		})
	})

	http.ListenAndServe(":8080", r)
}
