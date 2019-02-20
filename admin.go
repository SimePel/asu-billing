package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	ldap "gopkg.in/ldap.v3"
)

const (
	ldapServer = "ads.mc.asu.ru:3268"
)

var (
	admT = template.Must(template.New("adm").ParseGlob("templates/adm/*.html"))
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

	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
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
		log.Println(err)
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	session.Values["admin_logged"] = "true"
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// CorrectedUsers is a slice of CorrectedUser
type CorrectedUsers struct {
	Users []CorrectedUser
}

func adminIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	client, err := mongo.Connect(context.Background(), "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	coll := client.Database("billing").Collection("users")
	showType := r.URL.Query().Get("type")
	cur, err := getAppropriateCursor(coll, showType)
	if err != nil {
		log.Fatal(err)
	}

	users := make([]User, 0)
	user := User{}
	for cur.Next(context.Background()) {
		err := cur.Decode(&user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	cur.Close(context.Background())

	var correctedUsers CorrectedUsers
	for _, user := range users {
		correctedUsers.Users = append(correctedUsers.Users, CorrectedUser{
			Tariff:       tariffStringRepr(user.Tariff),
			Money:        user.Money,
			Name:         user.Name,
			Login:        user.Login,
			InIP:         user.InIP,
			ExtIP:        user.ExtIP,
			PaymentsEnds: formatTime(user.PaymentsEnds),
		})
	}
	admT.ExecuteTemplate(w, "index", correctedUsers)
}

func getAppropriateCursor(coll *mongo.Collection, showType string) (*mongo.Cursor, error) {
	if showType == "wired" {
		return coll.Find(context.Background(), bson.D{
			{Key: "tariff", Value: bson.D{
				{Key: "$in", Value: bson.A{1, 2}}},
			},
		})
	}
	if showType == "wireless" {
		return coll.Find(context.Background(), bson.D{
			{Key: "tariff", Value: 3},
		})
	}
	if showType == "active" {
		return coll.Find(context.Background(), bson.D{
			{Key: "payments_ends", Value: bson.D{
				{Key: "$gte", Value: "new Date()"},
			}},
		})
	}
	if showType == "inactive" {
		return coll.Find(context.Background(), bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "payments_ends", Value: nil}},
				bson.D{
					{Key: "payments_ends", Value: bson.D{
						{Key: "$lt", Value: "new Date()"},
					}},
				},
			}},
		})
	}
	return coll.Find(context.Background(), bson.D{})
}

func adminLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	session.Values["admin_logged"] = "false"
	session.Save(r, w)
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func userInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "false" || session.Values["admin_logged"] == nil {
		http.Redirect(w, r, "/admin-login", http.StatusFound)
		return
	}

	client, err := mongo.Connect(context.Background(), "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	user := User{}
	login := r.URL.Query().Get("login")
	filter := bson.D{{Key: "login", Value: login}}
	coll := client.Database("billing").Collection("users")
	err = coll.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	admT.ExecuteTemplate(w, "user-info", CorrectedUser{
		Tariff:       tariffStringRepr(user.Tariff),
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		PaymentsEnds: formatTime(user.PaymentsEnds),
	})
}

func formatTime(t time.Time) string {
	if t.Unix() < 0 {
		return "Оплата еще не производилась"
	}
	return t.Format("2.01.2006")
}

func newUserForm(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "admin")
	if session.Values["admin_logged"] == "true" {
		admT.ExecuteTemplate(w, "new-user-form", nil)
		return
	}
	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func addNewUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		log.Println("form parsing: ", err)
		http.Error(w, "Problems with fetching your data from the form. Please try again", http.StatusInternalServerError)
		return
	}
}
