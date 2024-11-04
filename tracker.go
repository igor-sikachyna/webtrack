package main

import (
	"fmt"
	"time"
	"webtrack/autoini"
	"webtrack/webfetch"
)

func trackerThread(config QueryConfig, stopRequest chan any, threadStopResponse chan any) {
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
					fmt.Printf("Failed to find the requested section on the page %v: %v", config.Url, err)
				} else {
					fmt.Println(res, err)
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

func StartTrackers(configs []string, stopRequest chan any, stopResponse chan any) {
	go func() {
		var stopChannels = []chan any{}
		for _, configPath := range configs {
			var threadStopResponse = make(chan any)
			stopChannels = append(stopChannels, threadStopResponse)
			var config = autoini.ReadIni[QueryConfig](configPath)
			go trackerThread(config, stopRequest, threadStopResponse)
		}
		// Await all channels to terminate
		for _, c := range stopChannels {
			<-c
		}
		close(stopResponse)
	}()
}
