package autoini

import (
	"errors"
	"reflect"
	"unicode"

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

type implementationChecker interface{}

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

type Setter func(value reflect.Value, valueInConfig *ini.Key) (err error)
type DefaultHandler[Implementation implementationChecker] func(def Implementation, key string, value reflect.Value) (err error)

func setGenericKey[T Configurable, Implementation implementationChecker](result T, key string, value reflect.Value, cfg *ini.File, setter Setter, defaultHandler DefaultHandler[Implementation]) (err error) {
	valueInConfig, err := getKey(key, isOptional(result, key), cfg)
	if err != nil {
		return err
	}

	if valueInConfig != nil {
		err = setter(value, valueInConfig)
	} else if def, ok := any(result).(Implementation); ok {
		err = defaultHandler(def, key, value)
	} else {
		return errors.New("did not find a value or default value for key: " + key)
	}
	return
}

func setStringKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	return setGenericKey(result, key, value, cfg,
		func(value reflect.Value, valueInConfig *ini.Key) (err error) {
			value.SetString(valueInConfig.String())
			return
		},
		func(def ImplementsDefaultString, key string, value reflect.Value) (err error) {
			value.SetString(def.DefaultString(key))
			return
		},
	)
}

func setIntKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	return setGenericKey(result, key, value, cfg,
		func(value reflect.Value, valueInConfig *ini.Key) (err error) {
			val, err := valueInConfig.Int64()
			if err != nil {
				return errors.New("config file key " + key + " is not an int")
			}
			value.SetInt(val)
			return
		},
		func(def ImplementsDefaultInt, key string, value reflect.Value) (err error) {
			value.SetInt(int64(def.DefaultInt(key)))
			return
		},
	)
}

func setBoolKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) (err error) {
	return setGenericKey(result, key, value, cfg,
		func(value reflect.Value, valueInConfig *ini.Key) (err error) {
			val, err := valueInConfig.Bool()
			if err != nil {
				return errors.New("config file key " + key + " is not a bool")
			}
			value.SetBool(val)
			return
		},
		func(def ImplementsDefaultBool, key string, value reflect.Value) (err error) {
			value.SetBool(def.DefaultBool(key))
			return
		},
	)
}

func ReadIni[T Configurable](path string) (result T, err error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return result, err
	}

	configReflection := reflect.ValueOf(&result)
	typeOfConfig := reflect.Indirect(configReflection).Type()

	for i := 0; i < reflect.Indirect(configReflection).NumField(); i++ {
		var fieldName = typeOfConfig.Field(i).Name
		if len(fieldName) == 0 || !unicode.IsUpper(rune(fieldName[0])) {
			return result, errors.New("field does not exist or is not exported: " + fieldName)
		}
		var configValueType = reflect.ValueOf(reflect.Indirect(configReflection).Field(i).Interface()).Type()

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
