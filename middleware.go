package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

func accessLog(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		since := time.Now()

		h(w, r, ps)

		var f *os.File
		f, err := os.OpenFile("logs/access.log", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			f, err = os.Create("logs/access.log")
			if err != nil {
				log.Printf("could not create file. skipping: %v", err)
				return
			}
		}

		fmt.Fprintf(f, "%v. Host: %v. Request: %v. Method: %v. Lead time: %v\n", time.Now().Format(time.UnixDate), r.RemoteAddr, r.RequestURI, r.Method, time.Now().Sub(since))
	}
}
