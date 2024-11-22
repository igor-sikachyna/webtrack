package autoini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SimpleValueNotExported struct {
	value bool
}

type SimpleValue struct {
	Value bool
}

type Values struct {
	value1 bool
}

func (v Values) Optional(key string) bool {
	return false
}

func TestReadIniInvalid(t *testing.T) {
	var assert = assert.New(t)

	var ini1, err = ReadIni[SimpleValueNotExported]("./test/a.ini")
	assert.NotEqual(nil, err, "Was able to read an invalid ini file")
	var errorString = err.Error()
	assert.Contains(errorString, "key-value delimiter not found")
	assert.Equal(false, ini1.value, "Incorrect value returned")

	ini1, err = ReadIni[SimpleValueNotExported]("./test/b.ini")
	assert.NotEqual(nil, err, "Was able to read an invalid ini file")
	errorString = err.Error()
	assert.Contains(errorString, "field does not exist or is not exported: value")
	assert.Equal(false, ini1.value, "Incorrect value returned")

	ini2, err := ReadIni[SimpleValue]("./test/b.ini")
	assert.NotEqual(nil, err, "Was able to read an ini file without the required value")
	errorString = err.Error()
	assert.Contains(errorString, "non-optional config file key not found: Value")
	assert.Equal(false, ini2.Value, "Incorrect value returned")
}

// func TestReadIniValid(t *testing.T) {
// 	var assert = assert.New(t)

// 	var ini = ReadIni[Values]()
// }
