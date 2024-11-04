package autoini

import (
	"gopkg.in/ini.v1"
	"log"
	"reflect"
)

func setStringKey(key string, value reflect.Value, cfg *ini.File) {
	valueInConfig := cfg.Section("").Key(key)
	if valueInConfig == nil {
		log.Fatalln("Config file ley not found: ", key)
	}
	value.SetString(valueInConfig.String())
}

func ReadIni[T any](path string) (result T) {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	configReflection := reflect.ValueOf(&result)
	typeOfConfig := reflect.Indirect(configReflection).Type()

	for i := 0; i < reflect.Indirect(configReflection).NumField(); i++ {
		var configValueType = reflect.ValueOf(reflect.Indirect(configReflection).Field(i).Interface()).Type()
		if configValueType.Name() == "string" {
			setStringKey(typeOfConfig.Field(i).Name, reflect.Indirect(configReflection).Field(i), cfg)
		}
	}

	return
}
