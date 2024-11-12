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

	// Drop the test collection before validating
	err = db.DropCollection("a")
	assert.Equal(nil, err, "Did not drop a collection")

	// MongoDB connector for Go allows to write to a collection which was not created using CreateCollection
	// This means that it is not really possible to fail a write operation due to non-existing collection
	err = db.Write("a", bson.D{{Key: "hello", Value: "world"}})
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

func TestGetLastDocument(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	// Drop the test collection before validating
	err = db.DropCollection("test")
	assert.Equal(nil, err, "Did not drop a collection")

	document, err := db.GetLastDocument("test", "none")
	assert.Equal(nil, err, "Did not return a document 1")
	assert.Equal(0, len(document), "Returned a non-empty document")

	err = db.Write("test", bson.D{{Key: "hello", Value: "a"}})
	assert.Equal(nil, err, "Did not write to a collection 1")

	document, err = db.GetLastDocument("test", "hello")
	assert.Equal(nil, err, "Did not return a document 2")
	assert.True(len(document) > 0, "Returned an empty document")
	rawDocument, err := BsonToRaw(document)
	assert.Equal(nil, err, "Failed to convert a document 1")
	assert.Equal("a", rawDocument.Lookup("hello").StringValue(), "Incorrect document value 1")

	err = db.Write("test", bson.D{{Key: "hello", Value: "c"}})
	assert.Equal(nil, err, "Did not write to a collection 2")

	err = db.Write("test", bson.D{{Key: "hello", Value: "b"}})
	assert.Equal(nil, err, "Did not write to a collection 3")

	document, err = db.GetLastDocument("test", "hello")
	assert.Equal(nil, err, "Did not return a document 3")
	assert.True(len(document) > 0, "Returned an empty document 2")
	rawDocument, err = BsonToRaw(document)
	assert.Equal(nil, err, "Failed to convert a document 2")
	assert.Equal("c", rawDocument.Lookup("hello").StringValue(), "Incorrect document value 2")

	document, err = db.GetLastDocument("test", "incorrect")
	assert.Equal(nil, err, "Did not return a document 4")
	assert.True(len(document) > 0, "Returned an empty document 3")
	rawDocument, err = BsonToRaw(document)
	assert.Equal(nil, err, "Failed to convert a document 3")
	assert.Equal("b", rawDocument.Lookup("hello").StringValue(), "Incorrect document value 3")
}

func TestGetLastDocumentFiltered(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	// Drop the test collection before validating
	err = db.DropCollection("test")
	assert.Equal(nil, err, "Did not drop a collection")

	db.Write("test", bson.D{{Key: "filter", Value: "ok"}, {Key: "hello", Value: "a"}})
	db.Write("test", bson.D{{Key: "filter", Value: "ok"}, {Key: "hello", Value: "b"}})
	db.Write("test", bson.D{{Key: "filter", Value: "notok"}, {Key: "hello", Value: "c"}})

	document, err := db.GetLastDocumentFiltered("test", "hello", bson.D{{Key: "filter", Value: "invalid"}})
	assert.Equal(nil, err, "Did not return a document for an invalid request")
	assert.Equal(0, len(document), "Returned a non-empty document")

	document, err = db.GetLastDocumentFiltered("test", "hello", bson.D{{Key: "filter", Value: "ok"}})
	assert.Equal(nil, err, "Did not return a document 1")
	assert.True(len(document) > 0, "Returned an empty document 1")
	rawDocument, err := BsonToRaw(document)
	assert.Equal(nil, err, "Failed to convert a document 1")
	assert.Equal("b", rawDocument.Lookup("hello").StringValue(), "Incorrect document value 1")

	document, err = db.GetLastDocumentFiltered("test", "hello", bson.D{{Key: "filter", Value: "notok"}})
	assert.Equal(nil, err, "Did not return a document 2")
	assert.True(len(document) > 0, "Returned an empty document 2")
	rawDocument, err = BsonToRaw(document)
	assert.Equal(nil, err, "Failed to convert a document 2")
	assert.Equal("c", rawDocument.Lookup("hello").StringValue(), "Incorrect document value 2")
}

func TestGetAllDocuments(t *testing.T) {
	var assert = assert.New(t)

	var db, err = NewMongoDB("mongodb://0.0.0.0:27017", "test")
	assert.Equal(nil, err, "Did not connect to a database")

	// Drop the test collection before validating
	err = db.DropCollection("test")
	assert.Equal(nil, err, "Did not drop a collection")

	documents, err := db.GetAllDocuments("test")
	assert.Equal(nil, err, "Did not return documents 1")
	assert.Equal(0, len(documents), "Returned a non-empty document list")

	db.Write("test", bson.D{{Key: "filter", Value: "ok"}, {Key: "hello", Value: "a"}})
	db.Write("test", bson.D{{Key: "filter", Value: "notok"}, {Key: "hello", Value: "b"}})

	documents, err = db.GetAllDocuments("test")
	assert.Equal(nil, err, "Did not return documents 2")
	assert.Equal(2, len(documents), "Incorrect documents count")
	rawDocument1, err := BsonToRaw(documents[0])
	assert.Equal(nil, err, "Failed to convert a document 1")
	rawDocument2, err := BsonToRaw(documents[1])
	assert.Equal(nil, err, "Failed to convert a document 2")
	assert.Equal("a", rawDocument1.Lookup("hello").StringValue(), "Incorrect document value 1-1")
	assert.Equal("ok", rawDocument1.Lookup("filter").StringValue(), "Incorrect document value 1-2")
	assert.Equal("b", rawDocument2.Lookup("hello").StringValue(), "Incorrect document value 2-1")
	assert.Equal("notok", rawDocument2.Lookup("filter").StringValue(), "Incorrect document value 2-2")
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
