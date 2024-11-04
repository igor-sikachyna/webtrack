package main

import (
	"fmt"
	"time"
	"webtrack/autoini"
	"webtrack/webfetch"
)

func trackerThread(config QueryConfig, stopRequest chan any, threadStopResponse chan any) {
	var fetcher = webfetch.NewFetcher()
	defer fetcher.Close()
	defer close(threadStopResponse)

	for {
		select {
		case <-stopRequest:
			return
		default:
			_, err := fetcher.FetchHtml(config.Url)

			if err != nil {
				// Not a critical issue, just log it
				fmt.Printf("Failed to query the page %v: %v", config.Url, err)
			} else {
				fmt.Println("Got page for ", config.Url)
				//fmt.Println(html)
			}
			time.Sleep(time.Duration(config.RequestIntervalSeconds) * time.Second)
			//return
		}
	}
}

func StartTrackers(configs []string, stopRequest chan any, stopResponse chan any) {
	go func() {
		var stopChannels []chan any = []chan any{}
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
