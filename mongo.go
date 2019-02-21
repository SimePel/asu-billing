package main

import (
	"log"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type CorrectedUsers struct {
	Users []CorrectedUser
}

func getUsersByType(t string) CorrectedUsers {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	coll := client.Database("billing").Collection("users")
	cur, err := getAppropriateCursor(coll, t)
	if err != nil {
		log.Fatal(err)
	}

	users := make([]User, 0)
	user := User{}
	for cur.Next(nil) {
		err := cur.Decode(&user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	cur.Close(nil)

	var correctedUsers CorrectedUsers
	for _, user := range users {
		correctedUsers.Users = append(correctedUsers.Users, CorrectedUser{
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

	return correctedUsers
}

func getAppropriateCursor(coll *mongo.Collection, showType string) (*mongo.Cursor, error) {
	if showType == "wired" {
		return coll.Find(nil, bson.D{
			{Key: "tariff", Value: bson.D{
				{Key: "$in", Value: bson.A{1, 2}}},
			},
		})
	}
	if showType == "wireless" {
		return coll.Find(nil, bson.D{
			{Key: "tariff", Value: 3},
		})
	}
	if showType == "active" {
		return coll.Find(nil, bson.D{
			{Key: "payments_ends", Value: bson.D{
				{Key: "$gte", Value: "new Date()"},
			}},
		})
	}
	if showType == "inactive" {
		return coll.Find(nil, bson.D{
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
	return coll.Find(nil, bson.D{})
}

func getUserDataByID(id int) CorrectedUser {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	user := User{}
	filter := bson.D{{Key: "_id", Value: id}}
	coll := client.Database("billing").Collection("users")
	err = coll.FindOne(nil, filter).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	return CorrectedUser{
		ID:           user.ID,
		Tariff:       tariffStringRepr(user.Tariff),
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		PaymentsEnds: formatTime(user.PaymentsEnds),
	}
}

func getUserDataByLogin(login string) CorrectedUser {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	user := User{}
	coll := client.Database("billing").Collection("users")
	err = coll.FindOne(nil, bson.M{"login": login}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}

	return CorrectedUser{
		ID:           user.ID,
		Tariff:       tariffStringRepr(user.Tariff),
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		PaymentsEnds: formatTime(user.PaymentsEnds),
	}
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

func formatTime(t time.Time) string {
	if t.Unix() < 0 {
		return "Оплата еще не производилась"
	}
	return t.Format("2.01.2006")
}
