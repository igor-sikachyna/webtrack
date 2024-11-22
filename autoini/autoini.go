package autoini

import (
	"errors"
	"reflect"

	"gopkg.in/ini.v1"
)

type Configurable interface{}

type ImplementsOptional interface {
	Optional(key string) bool
}

type ImplementsDefaultString interface {
	DefaultString(key string) string
}

type ImplementsDefaultInt interface {
	DefaultInt(key string) int
}

type ImplementsDefaultBool interface {
	DefaultBool(key string) bool
}

type ImplementsPostInit interface {
	// Must be implemented as a pointer receiver
	PostInit() (err error)
}

func isOptional[T Configurable](result T, key string) bool {
	if def, ok := any(result).(ImplementsOptional); ok {
		return def.Optional(key)
	}
	return false
}

func getKey(key string, optional bool, cfg *ini.File) (valueInConfig *ini.Key, err error) {
	valueExists := cfg.Section("").HasKey(key)
	if !valueExists {
		if !optional {
			return valueInConfig, errors.New("non-optional config file key not found: " + key)
		}
	} else {
		valueInConfig = cfg.Section("").Key(key)
	}
	return
}

func setStringKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	valueInConfig, err := getKey(key, isOptional(result, key), cfg)
	if err != nil {
		return err
	}

	if valueInConfig != nil {
		value.SetString(valueInConfig.String())
	} else if def, ok := any(result).(ImplementsDefaultString); ok {
		value.SetString(def.DefaultString(key))
	} else {
		return errors.New("did not find a value or default value for key: " + key)
	}
	return
}

func setIntKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	valueInConfig, err := getKey(key, isOptional(result, key), cfg)
	if err != nil {
		return err
	}

	if valueInConfig != nil {
		val, err := valueInConfig.Int64()
		if err != nil {
			return errors.New("config file key " + key + " is not an int")
		}
		value.SetInt(val)
	} else if def, ok := any(result).(ImplementsDefaultInt); ok {
		value.SetInt(int64(def.DefaultInt(key)))
	} else {
		return errors.New("did not find a value or default value for key: " + key)
	}
	return
}

func setBoolKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	valueInConfig, err := getKey(key, isOptional(result, key), cfg)
	if err != nil {
		return err
	}

	if valueInConfig != nil {
		val, err := valueInConfig.Bool()
		if err != nil {
			return errors.New("config file key " + key + " is not a bool")
		}
		value.SetBool(val)
	} else if def, ok := any(result).(ImplementsDefaultBool); ok {
		value.SetBool(def.DefaultBool(key))
	} else {
		return errors.New("did not find a value or default value for key: " + key)
	}
	return
}

func ReadIni[T Configurable](path string) (result T, err error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return result, err
	}

	configReflection := reflect.ValueOf(&result)
	typeOfConfig := reflect.Indirect(configReflection).Type()

	for i := 0; i < reflect.Indirect(configReflection).NumField(); i++ {
		var configValueType = reflect.ValueOf(reflect.Indirect(configReflection).Field(i).Interface()).Type()
		var fieldName = typeOfConfig.Field(i).Name
		switch configValueType.Name() {
		case "string":
			err = setStringKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		case "int":
			err = setIntKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		case "int64":
			err = setIntKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		case "bool":
			err = setBoolKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		default:
			return result, errors.New("unsupported ini value type: " + configValueType.Name())
		}

		if err != nil {
			return
		}
	}

	// Convert to a pointer to support pointer receivers
	if def, ok := any(&result).(ImplementsPostInit); ok {
		err = def.PostInit()
		if err != nil {
			return
		}
	}

	return
}
