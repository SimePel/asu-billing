package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func adminAuthCheck(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		session, _ := store.Get(r, "admin")
		if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
			http.Redirect(w, r, "/admin-login", http.StatusFound)
			return
		}
		h(w, r, ps)
	}
}
