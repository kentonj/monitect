package storage

import (
	"context"
	"fmt"
	"log"
	"monitect/conf"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	client *mongo.Client
	db     string
}

func (m *MongoDB) Ping() {
	err := m.client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Ping pong!")
	}
}

func Connect(config *conf.Config) *MongoDB {
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
	mongoDB := MongoDB{client: client, db: "db"}
	// do a ping check before returning
	mongoDB.Ping()
	return &mongoDB
}

func (m *MongoDB) Db() *mongo.Database {
	// return database object with parameterized database name
	return m.client.Database(m.db)
}

func (m *MongoDB) Close() {
	// defer this immediately after calling Connect()
	fmt.Println("Closing mongodb connection")
	if err := m.client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}

func (m *MongoDB) InsertOne(collectionName string, document interface{}) *mongo.InsertOneResult {
	// 2 second timeout for inserting one record
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := m.Db().Collection(collectionName)
	res, err := collection.InsertOne(ctx, document)
	if err != nil {
		log.Fatalf("unable to insert object to collection %s: %e", collectionName, err)
	} else {
		log.Printf("inserted record: collection %s object id %s", collectionName, res.InsertedID)
	}
	return res
}

func (m *MongoDB) FindOneByQuery(collectionName string, query bson.M) *mongo.SingleResult {
	// 2 second timeout for finding a record with a given query
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	collection := m.Db().Collection(collectionName)
	res := collection.FindOne(ctx, query)
	findErr := res.Err()
	if findErr == mongo.ErrNoDocuments {
		return nil
	} else if findErr != nil {
		log.Fatalf("Got an unrecoverable error %v", findErr)
	}
	return res
}
