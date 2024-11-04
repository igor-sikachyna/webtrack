package main

type Config struct {
	MongodbConnectionUrl  string
	DatabaseName          string
	VersionCollectionName string
}

func (cfg Config) Optional(key string) bool {
	// All values are mandatory
	return false
}
