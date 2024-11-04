package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"webtrack/autoini"
)

func main() {
	var config = autoini.ReadIni[Config]("config.ini")
	fmt.Println(config)

	var dir = "./queries"
	var stopRequest = make(chan any)
	var stopResponse = make(chan any)
	StartTrackers(ListFiles(dir), stopRequest, stopResponse)

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
