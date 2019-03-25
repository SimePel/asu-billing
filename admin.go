package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	ldap "gopkg.in/ldap.v3"
)

func newAdminTemplate() *template.Template {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	return template.Must(template.New("adm").Funcs(funcMap).ParseGlob("templates/adm/*.html"))
}

var (
	admT = newAdminTemplate()
)

func adminLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	admT.ExecuteTemplate(w, "login", nil)
}

func authAdmin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	login := r.FormValue("login")
	searchRequest := ldap.NewSearchRequest(
		"dc=mc,dc=asu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(memberOf=cn=billing,ou=groups,ou=vc,dc=mc,dc=asu,dc=ru)(samAccountName=%s))", login),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		url := fmt.Sprint("/admin-login?err=", err.Error())
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	session.Values["admin_logged"] = "true"
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func adminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// var name string
	t := r.FormValue("type")
	// if t == "name" {
	// 	name = r.FormValue("name")
	// }

	users, err := getUsersByType(t)
	if err != nil {
		log.Printf("could not get users by type=%v: %v", t, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	admT.ExecuteTemplate(w, "index", users)
}

func adminLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	session.Values["admin_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func userInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	user, err := getUserByID(id)
	if err != nil {
		log.Printf("could not get user by id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	admT.ExecuteTemplate(w, "user-info", user)
}

func userEditForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	user, err := getUserByID(id)
	if err != nil {
		log.Printf("could not get user by id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	admT.ExecuteTemplate(w, "edit-user-form", user)
}

func editUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := r.FormValue("name")
	login := r.FormValue("login")
	tariff := r.FormValue("tariff")
	phone := r.FormValue("phone")
	comment := r.FormValue("comment")

	err := updateUserData(id, name, login, tariff, phone, comment)
	if err != nil {
		log.Printf("could not update user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	admT.ExecuteTemplate(w, "new-user-form", nil)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.FormValue("name")
	login := r.FormValue("login") + "@stud.asu.ru"
	tariff := r.FormValue("tariff")
	phone := r.FormValue("phone")
	comment := r.FormValue("comment")

	moneyStr := r.FormValue("money")
	money := 0
	if moneyStr != "" {
		money, _ = strconv.Atoi(moneyStr)
	}

	user := User{
		Name:    name,
		Login:   login,
		Tariff:  tariffFromString(tariff),
		Phone:   phone,
		Comment: comment,
		Money:   money,
	}

	id, err := addUserToDB(user)
	if err != nil {
		log.Printf("could not add user into mongo with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	if money != 0 {
		err = addPaymentInfo(id, money)
		if err != nil {
			log.Printf("could not add payment info about user with id=%v: %v", id, err)
			http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
			return
		}
	}

	err = withdrawMoney(id)
	if err != nil {
		log.Printf("could not withdraw money from user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func deleteUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	err := deleteUserByID(id)
	if err != nil {
		log.Printf("could not delete user from mongo with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func payForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))

	user, err := getUserByID(id)
	if err != nil {
		log.Printf("could not get user by id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	admT.ExecuteTemplate(w, "payment", user)
}

func pay(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	moneyStr := r.FormValue("money")
	money, _ := strconv.Atoi(moneyStr)

	err := addMoney(id, money)
	if err != nil {
		log.Printf("could not add money to user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = addPaymentInfo(id, money)
	if err != nil {
		log.Printf("could not add payment info about user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	err = withdrawMoney(id)
	if err != nil {
		log.Printf("could not withdraw money from user with id=%v: %v", id, err)
		http.Error(w, "Что-то пошло не так", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func tariffFromString(s string) (t Tariff) {
	pieces := strings.Split(s, " ")
	t.ID, _ = strconv.Atoi(pieces[0])
	t.Name = pieces[1]
	t.Price, _ = strconv.Atoi(pieces[2])
	return t
}
