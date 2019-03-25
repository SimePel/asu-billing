package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func turnOffInactiveUsers() error {
	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return fmt.Errorf("could not connect to mongo: %v", err)
	}
	coll := client.Database("billing").Collection("users")

	ips, err := getDebtUsersIPs(coll)
	if err != nil {
		return fmt.Errorf("could not get ips of debt users: %v", err)
	}
	if ips == nil {
		return nil
	}

	err = disableUsers(coll, ips)
	if err != nil {
		return fmt.Errorf("could not disable users in mongo: %v", err)
	}

	err = removeUsersFromRouter(ips)
	if err != nil {
		return fmt.Errorf("could not block users ips on router: %v", err)
	}

	return nil
}

func getDebtUsersIPs(coll *mongo.Collection) ([]string, error) {
	cur, err := coll.Find(nil, bson.D{
		{Key: "payments_ends", Value: bson.D{
			{Key: "$lte", Value: time.Now()},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("could not find cursor when payments_ends < current date: %v", err)
	}

	var user User
	ips := make([]string, 0)
	for cur.Next(nil) {
		err := cur.Decode(&user)
		if err != nil {
			return nil, fmt.Errorf("could not decode data from mongo to user struct: %v", err)
		}
		ips = append(ips, user.InIP)
	}
	cur.Close(nil)

	return ips, nil
}

func disableUsers(coll *mongo.Collection, ips []string) error {
	for _, ip := range ips {
		_, err := coll.UpdateOne(nil,
			bson.D{
				{Key: "in_ip", Value: ip},
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
			return fmt.Errorf("could not update user active status: %v", err)
		}
	}

	return nil
}

func removeUsersFromRouter(ips []string) error {
	for _, ip := range ips {
		expectCMD := exec.Command("echo", ip)
		err := expectCMD.Run()
		if err != nil {
			return fmt.Errorf("could not run remove-expect cmd: %v", err)
		}
	}

	return nil
}

// func withdrawMoney(id int) error {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")
// 	user := User{}
// 	err = coll.FindOne(nil, bson.D{{Key: "_id", Value: id}}).Decode(&user)
// 	if err != nil {
// 		return fmt.Errorf("could not decode data from mongo to user struct: %v", err)
// 	}

// 	if user.Money < user.Tariff.Price {
// 		return nil
// 	}

// 	months := user.Money / user.Tariff.Price

// 	if !user.Active {
// 		paymentsEnds := time.Now().AddDate(0, months, 0)
// 		_, err := coll.UpdateOne(nil,
// 			bson.D{
// 				{Key: "_id", Value: user.ID},
// 			},
// 			bson.D{
// 				{Key: "$set", Value: bson.D{
// 					{Key: "payments_ends", Value: paymentsEnds},
// 					{Key: "active", Value: true},
// 				}},
// 				{Key: "$inc", Value: bson.D{
// 					{Key: "money", Value: -(user.Tariff.Price * months)},
// 				}},
// 			},
// 		)
// 		if err != nil {
// 			return fmt.Errorf("could not update \"payments_ends\" field: %v", err)
// 		}

// 		err = addUserIPToRouter(user.InIP)
// 		if err != nil {
// 			return fmt.Errorf("could not permit user's ip on router: %v", err)
// 		}

// 		return nil
// 	}

// 	paymentsEnds := user.PaymentsEnds.AddDate(0, months, 0)
// 	_, err = coll.UpdateOne(nil,
// 		bson.D{
// 			{Key: "_id", Value: user.ID},
// 		},
// 		bson.D{
// 			{Key: "$set", Value: bson.D{
// 				{Key: "payments_ends", Value: paymentsEnds},
// 			}},
// 			{Key: "$inc", Value: bson.D{
// 				{Key: "money", Value: -(user.Tariff.Price * months)},
// 			}},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("could not update \"payments_ends\" field: %v", err)
// 	}

// 	return nil
// }

func addUserIPToRouter(ip string) error {
	expectCMD := exec.Command("echo", "add "+ip)
	var out bytes.Buffer
	expectCMD.Stdout = &out
	err := expectCMD.Run()
	if err != nil {
		return fmt.Errorf("could not run add-expect cmd: %v", err)
	}

	fmt.Printf("%q", out)
	return nil
}

// func addPaymentInfo(id, money int) error {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")
// 	_, err = coll.UpdateOne(nil,
// 		bson.D{
// 			{Key: "_id", Value: id},
// 		},
// 		bson.D{
// 			{Key: "$push", Value: bson.D{
// 				{Key: "payments", Value: bson.D{
// 					{Key: "amount", Value: money},
// 					{Key: "last", Value: time.Now()},
// 				}},
// 			}},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("could not add payment info: %v", err)
// 	}

// 	return nil
// }

// func deleteUserFromMongo(id int) error {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")
// 	res := coll.FindOneAndDelete(nil, bson.D{{Key: "_id", Value: id}})

// 	var user User
// 	err = res.Decode(&user)
// 	if err != nil {
// 		return fmt.Errorf("could not decode data from mongo to user struct: %v", err)
// 	}

// 	coll = client.Database("billing").Collection("inIPs")
// 	_, err = coll.UpdateOne(nil,
// 		bson.D{
// 			{Key: "ip", Value: user.InIP},
// 		},
// 		bson.D{
// 			{Key: "$set", Value: bson.D{
// 				{Key: "used", Value: false},
// 			}},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("could not update ip used status: %v", err)
// 	}

// 	return nil
// }

func updateUserData(id int, name, login, tariff, phone, comment string) error {
	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return fmt.Errorf("could not connect to mongo: %v", err)
	}

	t := tariffFromString(tariff)
	coll := client.Database("billing").Collection("users")
	_, err = coll.UpdateOne(nil,
		bson.D{
			{Key: "_id", Value: id},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: name},
				{Key: "login", Value: login},
				{Key: "tariff", Value: bson.D{
					{Key: "id", Value: t.ID},
					{Key: "name", Value: t.Name},
					{Key: "price", Value: t.Price},
				}},
				{Key: "phone", Value: phone},
				{Key: "comment", Value: comment},
			}},
		},
	)
	if err != nil {
		return fmt.Errorf("could not update user fields: %v", err)
	}

	return nil
}

// func addUserIntoMongo(name, login, tariff, phone, comment string, money int) (int, error) {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return 0, fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")

// 	all, err := coll.CountDocuments(nil, bson.D{{}})
// 	if err != nil {
// 		return 0, fmt.Errorf("could not count documents: %v", err)
// 	}

// 	t := tariffFromString(tariff)
// 	inIP, err := getUnusedInIP(client)
// 	if err != nil {
// 		return 0, fmt.Errorf("could not get unused in_ip: %v", err)
// 	}
// 	_, err = coll.InsertOne(nil, bson.D{
// 		{Key: "_id", Value: int(all + 1)},
// 		{Key: "name", Value: name},
// 		{Key: "login", Value: login},
// 		{Key: "tariff", Value: bson.D{
// 			{Key: "id", Value: t.ID},
// 			{Key: "name", Value: t.Name},
// 			{Key: "price", Value: t.Price},
// 		}},
// 		{Key: "payments", Value: bson.A{}},
// 		{Key: "money", Value: money},
// 		{Key: "active", Value: false},
// 		{Key: "in_ip", Value: inIP},
// 		{Key: "ext_ip", Value: "82.200.46.10"}, // temporarily
// 		{Key: "phone", Value: phone},
// 		{Key: "comment", Value: comment},
// 	})
// 	if err != nil {
// 		return 0, fmt.Errorf("could not insert user data: %v", err)
// 	}

// 	return int(all + 1), nil
// }

// func getUnusedInIP(client *mongo.Client) (string, error) {
// 	coll := client.Database("billing").Collection("inIPs")

// 	var ip struct {
// 		IP   string `bson:"ip"`
// 		Used bool   `bson:"used"`
// 	}
// 	err := coll.FindOneAndUpdate(nil,
// 		bson.D{{Key: "used", Value: false}},
// 		bson.D{{Key: "$set", Value: bson.D{
// 			{Key: "used", Value: true},
// 		}}}).Decode(&ip)
// 	if err != nil {
// 		return "", fmt.Errorf("could not decode data from mongo to ip struct: %v", err)
// 	}

// 	return ip.IP, nil
// }

// func addMoneyToUser(id, money int) error {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")
// 	_, err = coll.UpdateOne(nil,
// 		bson.D{
// 			{Key: "_id", Value: id},
// 		},
// 		bson.D{
// 			{Key: "$inc", Value: bson.D{
// 				{Key: "money", Value: money},
// 			}},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("could not update money field: %v", err)
// 	}

// 	return nil
// }

// func getUsersByType(t, name string) ([]User, error) {
// 	cur, err := getAppropriateCursor(t, name)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get mongo.Cursor: %v", err)
// 	}

// 	users := make([]User, 0)
// 	user := User{}
// 	for cur.Next(nil) {
// 		err := cur.Decode(&user)
// 		if err != nil {
// 			return nil, fmt.Errorf("could not decode data from mongo to user struct: %v", err)
// 		}
// 		users = append(users, user)
// 	}
// 	cur.Close(nil)

// 	return users, nil
// }

// func getAppropriateCursor(showType, name string) (*mongo.Cursor, error) {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return nil, fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	coll := client.Database("billing").Collection("users")
// 	if showType == "name" {
// 		return coll.Find(nil, bson.D{
// 			{Key: "$text", Value: bson.D{
// 				{Key: "$search", Value: name},
// 			}},
// 		})
// 	}
// 	if showType == "wired" {
// 		return coll.Find(nil, bson.D{
// 			{Key: "tariff.id", Value: bson.D{
// 				{Key: "$in", Value: bson.A{1, 2}},
// 			}},
// 		})
// 	}
// 	if showType == "wireless" {
// 		return coll.Find(nil, bson.D{
// 			{Key: "tariff.id", Value: 3},
// 		})
// 	}
// 	if showType == "active" {
// 		return coll.Find(nil, bson.D{
// 			{Key: "active", Value: true},
// 		})
// 	}
// 	if showType == "inactive" {
// 		return coll.Find(nil, bson.D{
// 			{Key: "active", Value: false},
// 		})
// 	}

// 	return coll.Find(nil, bson.D{})
// }

// func getUserByID(id int) (User, error) {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return User{}, fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	user := User{}
// 	filter := bson.D{{Key: "_id", Value: id}}
// 	coll := client.Database("billing").Collection("users")
// 	err = coll.FindOne(nil, filter).Decode(&user)
// 	if err != nil {
// 		return User{}, fmt.Errorf("could not decode data from mongo to user struct: %v", err)
// 	}

// 	return user, nil
// }

// func getUserByLogin(login string) (User, error) {
// 	client, err := mongo.Connect(nil, options.Client().ApplyURI("mongodb://localhost:27017"))
// 	if err != nil {
// 		return User{}, fmt.Errorf("could not connect to mongo: %v", err)
// 	}

// 	user := User{}
// 	coll := client.Database("billing").Collection("users")
// 	err = coll.FindOne(nil, bson.M{"login": login}).Decode(&user)
// 	if err != nil {
// 		return User{}, fmt.Errorf("could not decode data from mongo to user struct: %v", err)
// 	}

// 	return user, nil
// }

func formatTime(t time.Time) string {
	if t.Unix() < 0 {
		return ""
	}
	return t.Format("2.01.2006")
}
