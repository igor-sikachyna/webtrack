package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
