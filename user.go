package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	ldap "gopkg.in/ldap.v3"
)

var (
	usrT = template.Must(template.New("usr").ParseGlob("templates/usr/*.html"))
)

// User is an instance of users collection from mongodb
type User struct {
	ID           int       `bson:"_id"`
	Money        int       `bson:"money"`
	Tariff       int       `bson:"tariff"`
	Name         string    `bson:"name"`
	Login        string    `bson:"login"`
	InIP         string    `bson:"in_ip"`
	ExtIP        string    `bson:"ext_ip"`
	PaymentsEnds time.Time `bson:"payments_ends,omitempty"`
}

// CorrectedUser needs to print appropriate information about user
type CorrectedUser struct {
	ID           int
	Money        int
	Name         string
	Login        string
	Tariff       string
	InIP         string
	ExtIP        string
	PaymentsEnds string
}

func userIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "false" || session.Values["user_logged"] == nil {
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	client, err := mongo.Connect(context.Background(), "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	user := User{}
	flashes := session.Flashes()
	filter := bson.M{"login": flashes[len(flashes)-1].(string)}
	coll := client.Database("billing").Collection("users")
	err = coll.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	usrT.ExecuteTemplate(w, "index", CorrectedUser{
		ID:           user.ID,
		Tariff:       tariffStringRepr(user.Tariff),
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		PaymentsEnds: formatTime(user.PaymentsEnds),
	})
}

func tariffStringRepr(tariff int) string {
	if tariff == 1 {
		return "Базовый30мб (300р)"
	}
	if tariff == 2 {
		return "Базовый30мб+IP (400р)"
	}
	if tariff == 3 {
		return "Беспроводной (200р)"
	}
	return "Неизвестный тариф"
}

func userLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}
	usrT.ExecuteTemplate(w, "login", nil)
}

func authUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	if session.Values["user_logged"] == "true" {
		http.Redirect(w, r, "/user", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}

	login := r.FormValue("login")
	pieces := strings.Split(login, "\\")
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("dc=%s,dc=asu,dc=ru", pieces[0]),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(samAccountName=%s)", pieces[1]),
		[]string{"dn"},
		nil,
	)

	err := ldapAuth(w, r, searchRequest)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/user-login", http.StatusFound)
		return
	}

	session.Values["user_logged"] = "true"
	session.AddFlash(pieces[1] + getRightPostfix(pieces[0]))
	session.Save(r, w)
	http.Redirect(w, r, "/user", http.StatusFound)
}

func getRightPostfix(domain string) string {
	if domain == "stud" {
		return "@stud.asu.ru"
	}
	if domain == "mc" {
		return "@mc.asu.ru"
	}
	return "Неизвестный домен"
}

func userLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "user")
	session.Values["user_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/user-login", http.StatusFound)
}
