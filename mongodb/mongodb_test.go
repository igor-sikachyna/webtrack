package mongodb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewMongoDBValid(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Returned an error 1")
	assert.NotEqual(nil, db.client, "Did not return a proper client")
	assert.NotEqual(nil, db.database, "Did not return a proper database")

	db2, err := NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Returned an error 2")
	assert.NotEqual(nil, db2.client, "Did not return a second proper client")
	assert.NotEqual(nil, db2.database, "Did not return a second proper database")
	assert.NotEqual(db.client, db2.client, "Did not return a unique client")
	assert.NotEqual(db.database, db2.database, "Did not return a unique database")
}

func TestNewMongoDBInvalid(t *testing.T) {
	var assert = assert.New(t)

	var _, err = NewMongoDB("invalid://0.0.0.0:27017", "test")
	assert.NotEqual(nil, err, "Did not return an error 1")

	_, err = NewMongoDB("mongodb://0.0.0.0:9999", "test")
	assert.NotEqual(nil, err, "Did not return an error 2")

	// At this stage the database is not created, so it is not possible to validate the name
	// _, err = NewMongoDB("mongodb://0.0.0.0:27017", strings.Repeat("a", 10000))
	// assert.NotEqual(nil, err, "Did not return an error 3")
}

func TestDisconnect(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	err = db.Disconnect()
	assert.Equal(nil, err, "Returned an error on proper disconnect")

	var db2 = MongoDB{}
	err = db2.Disconnect()
	assert.NotEqual(nil, err, "Did not return an error for improper db disconnect")
}

func TestCreateCollection(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	err = db.CreateCollection("a")
	assert.Equal(nil, err, "Did not create a collection")

	err = db.CreateCollection("a")
	assert.Equal(nil, err, "Did not skip an existing collection")

	err = db.CreateCollection(strings.Repeat("a", 256))
	assert.NotEqual(nil, err, "Was able to use a long collection name")

	err = (&MongoDB{}).CreateCollection("a")
	assert.NotEqual(nil, err, "Was able to use an initialized database")
}

func TestWrite(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	// MongoDB connector for Go allows to write to a collection which was not created using CreateCollection
	// This means that it is not really possible to fail a write operation due to non-existing collection
	err = db.Write("invalid", bson.D{{Key: "hello", Value: "world"}})
	assert.Equal(nil, err, "Did not write to a collection 1")

	// Empty objects are also accepted
	err = db.Write("a", bson.D{})
	assert.Equal(nil, err, "Did not write to a collection 2")

	// Using a custom _id is also accepted
	err = db.Write("a", bson.D{{Key: "_id", Value: "z"}})
	assert.Equal(nil, err, "Did not write to a collection 3")

	err = (&MongoDB{}).Write("a", bson.D{})
	assert.NotEqual(nil, err, "Was able to use an initialized database")
}

func TestDropCollection(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	err = db.Write("a", bson.D{})
	assert.Equal(nil, err, "Did not write to a collection")
	err = db.DropCollection("a")
	assert.Equal(nil, err, "Did not drop a collection")
	documents, err := db.GetAllDocuments("a")
	// MongoDB allows to read a dropped collection, but it should be empty
	assert.Equal(nil, err, "Was not able to read from a dropped collection")
	assert.Equal(0, len(documents), "Was able to find entries in a dropped collection")

	err = (&MongoDB{}).DropCollection("a")
	assert.NotEqual(nil, err, "Was able to use an initialized database")
}
