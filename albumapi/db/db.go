package db

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const uri = "mongodb://admin:admin@localhost:27017/admin?authenticationDatabase=admin&connect=direct"

var MongoClient *mongo.Client

func ConnectDb() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Fatal error connecting to the db: %v", err)
	}
	log.RegisterExitHandler(func() {
		errShutdown := client.Disconnect(context.TODO())
		if errShutdown != nil {
			log.Fatalf("Error when shutting down mongodb connection: %s", errShutdown.Error())
		}
	})
	// If the connection is established, wait for ping for no more than 5 sec
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to connect %s. %v", uri, err)

	}
	MongoClient = client
	log.Infof("MongoDb was connected successfully for URL: %s", uri)
}
