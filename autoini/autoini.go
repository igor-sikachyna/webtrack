package autoini

import (
	"gopkg.in/ini.v1"
	"log"
	"reflect"
)

type Configurable interface {
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
	PostInit()
}

func getKey(key string, optional bool, cfg *ini.File) (valueInConfig *ini.Key) {
	valueExists := cfg.Section("").HasKey(key)
	if !valueExists {
		if !optional {
			log.Fatalln("Non-optional config file key not found: ", key)
		}
	} else {
		valueInConfig = cfg.Section("").Key(key)
	}
	return valueInConfig
}

func setStringKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) {
	valueInConfig := getKey(key, result.Optional(key), cfg)
	if valueInConfig != nil {
		value.SetString(valueInConfig.String())
	} else if def, ok := any(result).(ImplementsDefaultString); ok {
		value.SetString(def.DefaultString(key))
	}
}

func setIntKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) {
	valueInConfig := getKey(key, result.Optional(key), cfg)
	if valueInConfig != nil {
		val, err := valueInConfig.Int64()
		if err != nil {
			log.Fatalf("Config file key %v is not an int", key)
		}
		value.SetInt(val)
	} else if def, ok := any(result).(ImplementsDefaultInt); ok {
		value.SetInt(int64(def.DefaultInt(key)))
	}
}

func setBoolKey[T Configurable](result T, key string, value reflect.Value, cfg *ini.File) {
	valueInConfig := getKey(key, result.Optional(key), cfg)
	if valueInConfig != nil {
		val, err := valueInConfig.Bool()
		if err != nil {
			log.Fatalf("Config file key %v is not a bool", key)
		}
		value.SetBool(val)
	} else if def, ok := any(result).(ImplementsDefaultBool); ok {
		value.SetBool(def.DefaultBool(key))
	}
}

func ReadIni[T Configurable](path string) (result T) {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	configReflection := reflect.ValueOf(&result)
	typeOfConfig := reflect.Indirect(configReflection).Type()

	for i := 0; i < reflect.Indirect(configReflection).NumField(); i++ {
		var configValueType = reflect.ValueOf(reflect.Indirect(configReflection).Field(i).Interface()).Type()
		var fieldName = typeOfConfig.Field(i).Name
		switch configValueType.Name() {
		case "string":
			setStringKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		case "int":
			setIntKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		case "bool":
			setBoolKey(result, fieldName, reflect.Indirect(configReflection).Field(i), cfg)
		default:
			log.Fatalf("Unsupported ini value type: %v", configValueType.Name())
		}
	}

	// Convert to a pointer to support pointer receivers
	if def, ok := any(&result).(ImplementsPostInit); ok {
		def.PostInit()
	}

	return
}
