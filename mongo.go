package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func addUserIntoMongo(name, login string, tariff, money int) error {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		return fmt.Errorf("could not connect to mongo: %v", err)
	}

	coll := client.Database("billing").Collection("users")

	all, err := coll.CountDocuments(nil, bson.D{{}})
	if err != nil {
		return fmt.Errorf("could not count documents: %v", err)
	}

	_, err = coll.InsertOne(nil, bson.D{
		{Key: "_id", Value: int(all + 1)},
		{Key: "name", Value: name},
		{Key: "login", Value: login},
		{Key: "tariff", Value: tariff},
		{Key: "money", Value: money},
		{Key: "active", Value: false},
		{Key: "in_ip", Value: getUnusedInIP(client)},
		{Key: "ext_ip", Value: "82.200.46.10"}, // temporarily
	})
	if err != nil {
		return fmt.Errorf("could not insert: %v", err)
	}

	return nil
}

func getUnusedInIP(client *mongo.Client) string {
	coll := client.Database("billing").Collection("inIPs")

	var ip struct {
		IP   string `bson:"ip"`
		Used bool   `bson:"used"`
	}
	err := coll.FindOne(nil, bson.D{{Key: "used", Value: false}}).Decode(&ip)
	if err != nil {
		log.Fatal(err)
	}

	return ip.IP
}

func addMoneyToUser(id, money int) {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	coll := client.Database("billing").Collection("users")
	_, err = coll.UpdateOne(nil,
		bson.D{
			{Key: "_id", Value: id},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "money", Value: money},
			}},
		},
	)
	if err != nil {
		log.Fatal("could not update money field", err)
	}
}

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
			Active:       user.Active,
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
			{Key: "active", Value: true},
		})
	}
	if showType == "inactive" {
		return coll.Find(nil, bson.D{
			{Key: "active", Value: false},
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
		Active:       user.Active,
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
		Active:       user.Active,
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
