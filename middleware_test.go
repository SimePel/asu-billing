package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{\"id\": 1}")
	})
	ts := httptest.NewServer(jsonContentType(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/json")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	assert.Equal(t, "{\"id\": 1}", string(b))
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
}

func TestCheckJWTtoken(t *testing.T) {
	r := chi.NewRouter()
	r.With(checkJWTtoken).Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "secret page")
	})
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "login page")
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "login page", string(b))
}
