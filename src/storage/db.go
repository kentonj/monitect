package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/kentonj/monitect/src/conf"
)

type MongoConnection struct {
	client *mongo.Client
	db     string
}

func (m *MongoConnection) Ping() {
	err := m.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Ping pong!")
	}
}

func Connect(config *conf.Config) MongoConnection {
	// return a mongoDB object which makes it easier to handle closing and the context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	credentials := options.Credential{
		Username: config.Mongo.Username,
		Password: config.Mongo.Password,
	}
	uri := fmt.Sprintf("mongodb://%s:%s", config.Mongo.Host, config.Mongo.Port)
	clientOpts := options.Client().ApplyURI(uri).SetAuth(credentials)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	// do a ping check before returning
	mongoDB := MongoConnection{client: client, db: config.Mongo.DB}
	mongoDB.Ping()
	return mongoDB
}

func (m *MongoConnection) Db() *mongo.Database {
	return m.client.Database(m.db)
}

func (m *MongoConnection) Close() {
	// defer this immediately after calling Connect()
	fmt.Println("Closing mongodb connection")
	if err := m.client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}
