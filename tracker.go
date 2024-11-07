package main

import (
	"errors"
	"fmt"
	"log"
	"time"
	"webtrack/autoini"
	"webtrack/mongodb"
	"webtrack/webfetch"
)

func trackerThread(config QueryConfig, mongo mongodb.MongoDB, stopRequest chan any, threadStopResponse chan any) {
	var fetcher = webfetch.NewFetcher(config.RequestBackend)
	defer fetcher.Close()
	defer close(threadStopResponse)

	for {
		select {
		case <-stopRequest:
			return
		default:
			html, err := fetcher.FetchHtml(config.Url)

			// Time delays properly by taking into account the request time itself
			var timeBefore = time.Now().UnixMilli()

			if err != nil {
				// Not a critical issue, just log it
				fmt.Printf("Failed to query the page %v: %v", config.Url, err)
			} else {
				var res, err = ExtractValueFromString(html, config.Before, config.After, config.AnyTag)
				if err != nil {
					fmt.Printf("Failed to find the requested section on the page %v: %v\n", config.Url, err)
				} else {
					if config.ResultType == "number" {
						res, err := ToNumber(res)
						fmt.Println(res, err)
					} else {
						fmt.Println(res)
					}
				}
			}

			var timeAfter = time.Now().UnixMilli()
			var sleepDuration = int64(config.RequestIntervalSeconds)*1000 - (timeAfter - timeBefore)
			if sleepDuration > 0 {
				time.Sleep(time.Duration(sleepDuration) * time.Millisecond)
			}
		}
	}
}

func StartTrackers(configs []string, mongo mongodb.MongoDB, stopRequest chan any, stopResponse chan any) (err error) {
	// Reserve the "versions" name since it is used for query versioning
	for _, configPath := range configs {
		var fileName = GetFileNameWithoutExtension(configPath)
		if fileName == "versions" {
			return errors.New("versions name is reserved")
		}
	}

	go func() {
		var stopChannels = []chan any{}
		for _, configPath := range configs {
			var threadStopResponse = make(chan any)
			stopChannels = append(stopChannels, threadStopResponse)
			var config = autoini.ReadIni[QueryConfig](configPath)
			config.Name = GetFileNameWithoutExtension(configPath)
			var err = mongo.CreateCollection(config.Name)
			if err != nil {
				log.Fatal(err)
			}
			go trackerThread(config, stopRequest, threadStopResponse)
		}
		// Await all channels to terminate
		for _, c := range stopChannels {
			<-c
		}
		close(stopResponse)
	}()

	return
}
