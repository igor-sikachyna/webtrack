package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToNumberValid(t *testing.T) {
	var assert = assert.New(t)

	var num, err = ToNumber("123")
	assert.Equal(nil, err, "Returned an error 1")
	assert.Equal(123.0, num, "Incorrect number 1")

	num, err = ToNumber("abc456def")
	assert.Equal(nil, err, "Returned an error 2")
	assert.Equal(456.0, num, "Incorrect number 2")

	num, err = ToNumber("12.2")
	assert.Equal(nil, err, "Returned an error 3")
	assert.Equal(12.2, num, "Incorrect number 3")

	num, err = ToNumber("ttt 12,300. zzz")
	assert.Equal(nil, err, "Returned an error 4")
	assert.Equal(12300.0, num, "Incorrect number 4")

	num, err = ToNumber("123 456")
	assert.Equal(nil, err, "Returned an error 5")
	assert.Equal(123456.0, num, "Incorrect number 5")
}

func TestToNumberInvalid(t *testing.T) {
	var assert = assert.New(t)

	var _, err = ToNumber("abc")
	assert.NotEqual(nil, err, "Did not return an error 1")

	_, err = ToNumber("1..2")
	assert.NotEqual(nil, err, "Did not return an error 2")
}
