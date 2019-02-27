package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func turnOffInactiveUsers() error {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		return fmt.Errorf("could not connect to mongo: %v", err)
	}

	coll := client.Database("billing").Collection("users")
	_, err = coll.UpdateMany(nil,
		bson.D{
			{Key: "payments_ends", Value: bson.D{
				{Key: "$lte", Value: time.Now()},
			}},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "active", Value: false},
			}},
			{Key: "$unset", Value: bson.D{
				{Key: "payments_ends", Value: nil},
			}},
		},
	)
	if err != nil {
		return fmt.Errorf("could not turn off users: %v", err)
	}

	return nil
}

func withdrawMoney(id int) error {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		return fmt.Errorf("could not connect to mongo: %v", err)
	}

	coll := client.Database("billing").Collection("users")
	user := User{}
	err = coll.FindOne(nil, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		return fmt.Errorf("could not find user by id: %v", err)
	}

	if user.Money < user.Tariff.Price {
		return nil
	}

	months := user.Money / user.Tariff.Price

	if !user.Active {
		paymentsEnds := time.Now().AddDate(0, months, 0)
		_, err := coll.UpdateOne(nil,
			bson.D{
				{Key: "_id", Value: user.ID},
			},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "payments_ends", Value: paymentsEnds},
					{Key: "active", Value: true},
				}},
				{Key: "$inc", Value: bson.D{
					{Key: "money", Value: -(user.Tariff.Price * months)},
				}},
			},
		)
		if err != nil {
			return fmt.Errorf("could not update \"payments_ends\" field: %v", err)
		}
		return nil
	}

	paymentsEnds := user.PaymentsEnds.AddDate(0, months, 0)
	_, err = coll.UpdateOne(nil,
		bson.D{
			{Key: "_id", Value: user.ID},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "payments_ends", Value: paymentsEnds},
			}},
			{Key: "$inc", Value: bson.D{
				{Key: "money", Value: -(user.Tariff.Price * months)},
			}},
		},
	)
	if err != nil {
		return fmt.Errorf("could not update \"payments_ends\" field: %v", err)
	}

	return nil
}

func addUserIntoMongo(name, login, tariff, phone, comment string, money int) (int, error) {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		return 0, fmt.Errorf("could not connect to mongo: %v", err)
	}

	coll := client.Database("billing").Collection("users")

	all, err := coll.CountDocuments(nil, bson.D{{}})
	if err != nil {
		return 0, fmt.Errorf("could not count documents: %v", err)
	}

	pieces := strings.Split(tariff, " ")
	tID, _ := strconv.Atoi(pieces[0])
	tName := pieces[1]
	tPrice, _ := strconv.Atoi(pieces[2])
	_, err = coll.InsertOne(nil, bson.D{
		{Key: "_id", Value: int(all + 1)},
		{Key: "name", Value: name},
		{Key: "login", Value: login},
		{Key: "tariff", Value: bson.D{
			{Key: "id", Value: tID},
			{Key: "name", Value: tName},
			{Key: "price", Value: tPrice},
		}},
		{Key: "money", Value: money},
		{Key: "active", Value: false},
		{Key: "in_ip", Value: getUnusedInIP(client)},
		{Key: "ext_ip", Value: "82.200.46.10"}, // temporarily
		{Key: "phone", Value: phone},
		{Key: "comment", Value: comment},
	})
	if err != nil {
		return 0, fmt.Errorf("could not insert: %v", err)
	}

	return int(all + 1), nil
}

func getUnusedInIP(client *mongo.Client) string {
	coll := client.Database("billing").Collection("inIPs")

	var ip struct {
		IP   string `bson:"ip"`
		Used bool   `bson:"used"`
	}
	err := coll.FindOneAndUpdate(nil,
		bson.D{{Key: "used", Value: false}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "used", Value: true},
		}}}).Decode(&ip)
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

func getUsersByType(t, name string) CorrectedUsers {
	cur, err := getAppropriateCursor(t, name)
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
			Tariff:       user.Tariff,
			Money:        user.Money,
			Name:         user.Name,
			Login:        user.Login,
			InIP:         user.InIP,
			ExtIP:        user.ExtIP,
			Phone:        user.Phone,
			Comment:      user.Comment,
			PaymentsEnds: formatTime(user.PaymentsEnds),
		})
	}

	return correctedUsers
}

func getAppropriateCursor(showType, name string) (*mongo.Cursor, error) {
	client, err := mongo.Connect(nil, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("could not connect to mongo", err)
	}

	coll := client.Database("billing").Collection("users")
	if showType == "name" {
		return coll.Find(nil, bson.D{
			{Key: "$text", Value: bson.D{
				{Key: "$search", Value: name},
			}},
		})
	}
	if showType == "wired" {
		return coll.Find(nil, bson.D{
			{Key: "tariff.id", Value: bson.D{
				{Key: "$in", Value: bson.A{1, 2}},
			}},
		})
	}
	if showType == "wireless" {
		return coll.Find(nil, bson.D{
			{Key: "tariff.id", Value: 3},
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
		Tariff:       user.Tariff,
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		Phone:        user.Phone,
		Comment:      user.Comment,
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
		Tariff:       user.Tariff,
		Money:        user.Money,
		Name:         user.Name,
		Login:        user.Login,
		InIP:         user.InIP,
		ExtIP:        user.ExtIP,
		Phone:        user.Phone,
		Comment:      user.Comment,
		PaymentsEnds: formatTime(user.PaymentsEnds),
	}
}

func formatTime(t time.Time) string {
	if t.Unix() < 0 {
		return ""
	}
	return t.Format("2.01.2006")
}
