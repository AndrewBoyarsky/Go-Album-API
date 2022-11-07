package users

import (
	"context"
	"fmt"

	"github.com/AndrewBoyarsky/albumapi/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	UserName string
	Password string
}

type UserRepo interface {
	GetById(id string) *User

	GetByUserName(name string) *User

	Save(u User) string
}

func NewUserRepo() UserRepo {
	repo := mongUserRepo{}
	repo.store = db.MongoClient.Database("mydb").Collection("users")
	return repo
}

type mongUserRepo struct {
	store *mongo.Collection
}

func (m mongUserRepo) GetById(id string) *User {
	result := m.store.FindOne(context.TODO(), bson.D{{"_id", parseIdFromString(id)}})
	if result.Err() != nil {
		return nil
	} else {
		user := User{}
		err := result.Decode(&user)
		if err != nil {
			panic(err)
		}
		return &user
	}
}

func (m mongUserRepo) GetByUserName(id string) *User {
	result := m.store.FindOne(context.TODO(), bson.D{{"username", id}})
	if result.Err() != nil {
		return nil
	} else {
		user := User{}
		err := result.Decode(&user)
		if err != nil {
			panic(err)
		}
		return &user
	}
}

func (m mongUserRepo) Save(u User) string {
	result, err := m.store.InsertOne(context.TODO(), u)
	if err != nil {
		panic(err)
	}
	return (result.InsertedID.(primitive.ObjectID)).Hex()
}

func parseIdFromString(id string) primitive.ObjectID {
	idp, parseError := primitive.ObjectIDFromHex(id)
	if parseError != nil {
		panic(fmt.Errorf("unable to parse mongodb id from: %s, reason: %s", id, parseError.Error()))
	}
	return idp
}
