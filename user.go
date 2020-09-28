package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
)

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	mysql := MySQL{db: r.Context().Value(dbCtxKey("db")).(*sql.DB)}
	users, err := mysql.GetAllUsers()
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Println("cannot encode json. ", err)
		w.Write([]byte("{}"))
	}
}

type userCtxKey string

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		id, _ := strconv.Atoi(userID)
		mysql := MySQL{db: r.Context().Value(dbCtxKey("db")).(*sql.DB)}
		user, err := mysql.GetUserByID(id)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey("user"), &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// тут нужно проверить, что в контексте юзер ест, а если нет, то упасть красиво
	user := r.Context().Value(userCtxKey("user")).(*User)
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("cannot encode json. ", err)
		w.Write([]byte("{}"))
	}
}

func unlimitUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtxKey("user")).(*User)

	// Срезаем букву 'П' и '-'
	err := unlimitUserPhysically("P" + user.Agreement[3:])
	if err != nil {
		log.Printf("cannot unlimit user on the router: %v", err)
		return
	}

	mysql := MySQL{db: initializeDB()}
	err = mysql.UnlimitUserByID(int(user.ID))
	if err != nil {
		log.Println(err)
	}
}

func unlimitUserPhysically(agreement string) error {
	expect := exec.Command("expect", "unlimit.exp", "user_"+agreement+"_108_in", "user_"+agreement+"_108_out", "class_user_"+agreement+"_108_in", "class_user_"+agreement+"_108_out")
	out, err := expect.CombinedOutput()
	if err != nil {
		log.Printf("got %v\n", string(out))
		return fmt.Errorf("cannot execute unlimit expect script: %v", err)
	}

	return nil
}

func limitUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtxKey("user")).(*User)

	// Срезаем букву 'П' и '-'
	err := limitUserPhysically(user.InnerIP, "P"+user.Agreement[3:])
	if err != nil {
		log.Printf("cannot limit user on the router: %v", err)
		return
	}

	mysql := MySQL{db: initializeDB()}
	err = mysql.LimitUserByID(int(user.ID))
	if err != nil {
		log.Println(err)
	}
}

func limitUserPhysically(ip, agreement string) error {
	expect := exec.Command("expect", "limit.exp", ip, "user_"+agreement+"_108_in", "user_"+agreement+"_108_out", "class_user_"+agreement+"_108_in", "class_user_"+agreement+"_108_out")
	out, err := expect.CombinedOutput()
	if err != nil {
		log.Printf("got %v\n", string(out))
		return fmt.Errorf("cannot execute limit expect script: %v", err)
	}

	return nil
}

func deactivateUser(w http.ResponseWriter, r *http.Request) {
	token, err := getJWTtokenFromCookies(r.Cookies())
	if err != nil {
		log.Println(err)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	admin := claims["login"].(string)

	user := r.Context().Value(userCtxKey("user")).(*User)
	mysql := MySQL{db: initializeDB()}
	err = mysql.DeactivateUserByID(int(user.ID), admin)
	if err != nil {
		log.Printf("cannot deactivate user with id=%v: %v", user.ID, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
	}
}

func activateUser(w http.ResponseWriter, r *http.Request) {
	token, err := getJWTtokenFromCookies(r.Cookies())
	if err != nil {
		log.Println(err)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	admin := claims["login"].(string)

	user := r.Context().Value(userCtxKey("user")).(*User)
	mysql := MySQL{db: initializeDB()}
	err = mysql.ActivateUserByID(int(user.ID), admin)
	if err != nil {
		log.Printf("cannot activate user with id=%v: %v", user.ID, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
	}
}

func archiveUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtxKey("user")).(*User)
	mysql := MySQL{db: initializeDB()}
	err := mysql.ArchiveUserByID(int(user.ID))
	if err != nil {
		log.Println(err)
	}
}

func restoreUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userCtxKey("user")).(*User)
	mysql := MySQL{db: initializeDB()}
	err := mysql.RestoreUserByID(int(user.ID))
	if err != nil {
		log.Println(err)
	}
}
