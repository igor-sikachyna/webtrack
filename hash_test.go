package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileHashValid(t *testing.T) {
	var assert = assert.New(t)

	var hash, err = GetFileHash("hash_test.go")
	assert.Equal(nil, err, "Returned an error 1")
	assert.Equal(64, len(hash), "Incorrect hash length 1")

	hash, err = GetFileHash("test/example.txt")
	assert.Equal(nil, err, "Returned an error 2")
	assert.Equal("a9a66978f378456c818fb8a3e7c6ad3d2c83e62724ccbdea7b36253fb8df5edd", hash, "Incorrect hash length 2")
}

func TestGetFileHashInvalid(t *testing.T) {
	var assert = assert.New(t)

	var _, err = GetFileHash("invalid.exe")
	assert.NotEqual(nil, err, "Did not return an error 1")
}
