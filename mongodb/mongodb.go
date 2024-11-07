package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDB(uri string, databaseName string) (result MongoDB, err error) {
	result.client, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = result.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return result, err
	}

	result.database = result.client.Database(databaseName)

	return result, err
}

func Connect(uri string) (client *mongo.Client, err error) {
	return mongo.Connect(options.Client().ApplyURI(uri))
}

func (m *MongoDB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func (m *MongoDB) CreateCollection(collection string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Ensure that we create only missing collection
	names, err := m.database.ListCollectionNames(ctx, bson.D{{"name", collection}})
	if err != nil {
		return err
	}
	if len(names) == 1 && names[0] == collection {
		return
	}

	err = m.database.CreateCollection(ctx, collection)
	return err
}

func (m *MongoDB) Write(collection string, data bson.D) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	mongoCollection := m.database.Collection(collection)

	_, err = mongoCollection.InsertOne(ctx, data)
	return err
}
