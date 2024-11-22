package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"webtrack/autoini"
	"webtrack/mongodb"
)

func main() {
	var config, err = autoini.ReadIni[Config]("config.ini")
	if err != nil {
		log.Fatal(err)
	}

	mongo, err := mongodb.NewMongoDB(config.MongodbConnectionUrl, config.DatabaseName)
	if err != nil {
		log.Fatal(err)
	}
	defer mongo.Disconnect()

	// Create the default versions collection
	err = mongo.CreateCollection(config.VersionCollectionName)
	if err != nil {
		log.Fatal(err)
	}

	var dir = "./queries"
	var stopRequest = make(chan any)
	var stopResponse = make(chan any)
	err = StartTrackers(ListIniFiles(dir), config, mongo, stopRequest, stopResponse)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("webtrack initialized. Waiting for termination...")
	awaitTermination()
	fmt.Println("Gracefully exiting...")
	close(stopRequest)
	<-stopResponse
}

func awaitTermination() {
	wait := make(chan any)

	go func() {
		c := make(chan os.Signal, 1) // Need to reserve a buffer of size 1, otherwise the notifier will be blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		close(wait)
	}()

	<-wait
}
