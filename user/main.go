package main

import (
	"log"
	"net/http"
)

func main() {
	r := newRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}
