package autoini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type SimpleValue struct {
	value bool
}

type Values struct {
	value1 bool
}

func (v Values) Optional(key string) bool {
	return false
}

func TestReadIniInvalid(t *testing.T) {
	var assert = assert.New(t)

	var ini, err = ReadIni[SimpleValue]("./test/a.ini")
	assert.NotEqual(nil, err, "Was able to read an invalid ini file")
	assert.Contains(err, "key-value delimiter not found")
	assert.Equal(false, ini.value, "Incorrect value returned")

	ini, err = ReadIni[SimpleValue]("./test/b.ini")
	assert.NotEqual(nil, err, "Was able to read an ini file without the required value")
	assert.Contains(err, "key-value delimiter not found")
	assert.Equal(false, ini.value, "Incorrect value returned")
}

// func TestReadIniValid(t *testing.T) {
// 	var assert = assert.New(t)

// 	var ini = ReadIni[Values]()
// }
