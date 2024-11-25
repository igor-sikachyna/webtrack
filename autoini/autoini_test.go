package autoini

import (
	"errors"
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
	Value1 bool
	Value2 int
	Value3 string
}

func (v Values) Optional(key string) bool {
	return false
}

type ValuesDefault struct {
	Value1 bool
	Value2 int
	Value3 string
}

func (v ValuesDefault) Optional(key string) bool {
	return true
}

func (v ValuesDefault) DefaultString(key string) string {
	return "string value"
}

func (v ValuesDefault) DefaultInt(key string) int {
	return 123
}

func (v ValuesDefault) DefaultBool(key string) bool {
	return false
}

func (v ValuesDefault) PostInit() (err error) {
	if v.Value2 == 42 {
		return errors.New("42 is not allowed")
	}

	return nil
}

func TestReadIniInvalid(t *testing.T) {
	var assert = assert.New(t)

	var ini1, err = ReadIni[SimpleValueNotExported]("./test/a.ini")
	assert.NotEqual(nil, err, "Was able to read an invalid ini file")
	assert.Contains(err.Error(), "key-value delimiter not found")
	assert.Equal(false, ini1.value, "Incorrect value returned")

	ini1, err = ReadIni[SimpleValueNotExported]("./test/b.ini")
	assert.NotEqual(nil, err, "Was able to read an invalid ini file")
	assert.Contains(err.Error(), "field does not exist or is not exported: value")
	assert.Equal(false, ini1.value, "Incorrect value returned")

	ini2, err := ReadIni[SimpleValue]("./test/b.ini")
	assert.NotEqual(nil, err, "Was able to read an ini file without the required value")
	assert.Contains(err.Error(), "non-optional config file key not found: Value")
	assert.Equal(false, ini2.Value, "Incorrect value returned")
}

func TestReadIniValid(t *testing.T) {
	var assert = assert.New(t)

	var ini, err = ReadIni[Values]("./test/c.ini")
	assert.Equal(nil, err, "Was not able to read a valid ini file")
	assert.Equal(true, ini.Value1, "Incorrect Value1 returned")
	assert.Equal(42, ini.Value2, "Incorrect Value2 returned")
	assert.Equal("hello world", ini.Value3, "Incorrect Value3 returned")
}

func TestReadIniOptionalImplementationsValid(t *testing.T) {
	var assert = assert.New(t)

	var ini, err = ReadIni[ValuesDefault]("./test/empty.ini")
	assert.Equal(nil, err, "Was not able to read a valid ini file")
	assert.Equal(false, ini.Value1, "Incorrect Value1 returned")
	assert.Equal(123, ini.Value2, "Incorrect Value2 returned")
	assert.Equal("string value", ini.Value3, "Incorrect Value3 returned")

	ini, err = ReadIni[ValuesDefault]("./test/c.ini")
	assert.NotEqual(nil, err, "PostInit did not return an error")
	assert.Contains(err.Error(), "42 is not allowed")
	assert.Equal(42, ini.Value2, "Incorrect new Value2 returned")
}
