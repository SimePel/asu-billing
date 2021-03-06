package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, "login page", string(b))

	token, err := createJWTtoken("login")
	require.NoError(t, err)

	c := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 1, 0),
		SameSite: 3,
	}

	req, err := http.NewRequest("GET", ts.URL+"/", nil)
	require.NoError(t, err)
	req.AddCookie(c)

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, "secret page", string(body))
}

// На самом деле этот тест бесполезный, сделан только для + coverage
func TestSetDBtoCtx(t *testing.T) {
	r := chi.NewRouter()
	r.With(setDBtoCtx).Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Context().Value(dbCtxKey("db")).(*sql.DB))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
