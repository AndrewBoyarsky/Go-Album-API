package albums

import (
	"context"
	"fmt"

	"github.com/AndrewBoyarsky/albumapi/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlbumRepo interface {
	GetById(ctx mongo.SessionContext, id string, userName string) *Album

	DeleteById(ctx mongo.SessionContext, id string, userName string) bool

	Save(ctx mongo.SessionContext, album Album, id string) string

	GetAll(ctx mongo.SessionContext, userName string) []Album
}

type mongoDbAlbumRepo struct {
	store *mongo.Collection
}

func NewAlbumRepo() AlbumRepo {
	repo := mongoDbAlbumRepo{}
	repo.store = db.MongoClient.Database("mydb").Collection("albums")
	return repo
}

func (m mongoDbAlbumRepo) GetById(ctx mongo.SessionContext, id string, userName string) *Album {

	idp := parseIdFromString(id)
	result := m.store.FindOne(getContext(ctx), bson.D{{"_id", idp}, {"userName", userName}})
	if result.Err() != nil {
		return nil
	} else {
		mapped := Album{}
		err := result.Decode(&mapped)
		if err != nil {
			panic(fmt.Errorf("unable to map mongodb document of album to actual struct: %s", err.Error()))
		}
		return &mapped
	}
}

func (m mongoDbAlbumRepo) DeleteById(ctx mongo.SessionContext, id string, userName string) bool {
	idp := parseIdFromString(id)
	result := m.store.FindOneAndDelete(getContext(ctx), bson.D{{"_id", idp}, {"userName", userName}})
	if result.Err() != nil {
		return false
	} else {
		return true
	}
}

func parseIdFromString(id string) primitive.ObjectID {
	idp, parseError := primitive.ObjectIDFromHex(id)
	if parseError != nil {
		panic(fmt.Errorf("unable to parse mongodb id from: %s, reason: %s", id, parseError.Error()))
	}
	return idp
}

func (m mongoDbAlbumRepo) Save(ctx mongo.SessionContext, album Album, id string) string {
	if id == "" {
		inserted, err := m.store.InsertOne(getContext(ctx), album)
		if err != nil {
			panic(err)
		}
		return inserted.InsertedID.(primitive.ObjectID).Hex()
	} else {
		idp := parseIdFromString(id)
		result := m.store.FindOneAndReplace(getContext(ctx), bson.M{"_id": idp, "userName": album.UserName}, album)
		if result.Err() != nil {
			return ""
		} else {
			return id
		}
	}
}

func (m mongoDbAlbumRepo) GetAll(ctx mongo.SessionContext, userName string) []Album {
	cursor, err := m.store.Find(getContext(ctx), bson.D{{"userName", userName}})
	if err != nil {
		panic(err)
	}
	result := []Album{}
	errMapping := cursor.All(getContext(ctx), &result)
	if errMapping != nil {
		panic(errMapping)
	}
	return result
}

func getContext(ctx mongo.SessionContext) context.Context {

	if ctx == nil {
		return context.TODO()
	} else {
		return ctx
	}
}
