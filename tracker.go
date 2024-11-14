package main

import (
	"errors"
	"fmt"
	"log"
	"time"
	"webtrack/autoini"
	"webtrack/mongodb"
	"webtrack/webfetch"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TrackedRecord struct {
	Timestamp int64
	Value     string
	Version   int64
}

type VersionRecord struct {
	Name    string
	Version int64
	Hash    string
}

func floatToNiceString(value float64) (res string) {
	res = fmt.Sprintf("%f", value)
	// Trim 0s past the decimal point
	var trimEnd = len(res)
	for i := len(res) - 1; i >= 0; i-- {
		if res[i] == '0' {
			trimEnd = i
		} else if res[i] == '.' {
			trimEnd = i
		} else {
			break
		}
	}

	return res[0:trimEnd]
}

func trackerThread(config QueryConfig, mongo mongodb.MongoDB, stopRequest chan any, threadStopResponse chan any) {
	var fetcher = webfetch.NewFetcher(config.RequestBackend)
	defer fetcher.Close()
	defer close(threadStopResponse)

	var lastValue = ""
	if config.OnlyIfDifferent {
		var lastDocument, err = mongo.GetLastDocument(config.Name, "timestamp")
		if lastDocument != nil {
			if err != nil {
				log.Fatal(err)
			}

			var decoded TrackedRecord
			err = lastDocument.Decode(&decoded)
			if err != nil {
				log.Fatal(err)
			}
			lastValue = decoded.Value
		}
	}

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
						var number float64
						number, err = ToNumber(res)
						if err != nil {
							fmt.Printf("Failed to convert %v to a number: %v", res, err)
						} else {
							res = floatToNiceString(number)
						}
					}

					// Respect the OnlyIfDifferent and OnlyIfUnique requirement
					var onlyIfDifferentPassed = (!config.OnlyIfDifferent || lastValue != res)
					// Small optimization: if the last record is the same as the current, then it is not necessary to search in MongoDB
					var onlyIfUniquePassed = onlyIfDifferentPassed
					if onlyIfUniquePassed && config.OnlyIfUnique {
						onlyIfUniquePassed = false
						existingDocument, err := mongo.GetLastDocumentFiltered(config.Name, "timestamp", bson.D{{Key: "value", Value: res}, {Key: "version", Value: config.Version}})
						if err != nil {
							fmt.Printf("Failed the search for an existing record in MongoDB: %v", err)
						} else if existingDocument != nil {
							onlyIfUniquePassed = true
						}
					}

					if err == nil && onlyIfDifferentPassed && onlyIfUniquePassed {
						var timestamp = time.Now().Unix()
						err = mongo.Write(config.Name, bson.D{{Key: "timestamp", Value: timestamp}, {Key: "value", Value: res}, {Key: "version", Value: config.Version}})
						if err != nil {
							fmt.Printf("Failed to write to MongoDB: %v", err)
						} else {
							fmt.Printf("Wrote to MongoDB collection %v at %v\n", config.Name, timestamp)
							lastValue = res
						}
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

func writeQueryVersion(mongo mongodb.MongoDB, globalConfig Config, queryName string, queryHash string, version int64) error {
	return mongo.Write(globalConfig.VersionCollectionName, bson.D{{Key: "name", Value: queryName}, {Key: "version", Value: version}, {Key: "hash", Value: queryHash}})
}

func checkQueryVersion(mongo mongodb.MongoDB, globalConfig Config, queryName string, configPath string) (int64, error) {
	document, err := mongo.GetLastDocumentFiltered(globalConfig.VersionCollectionName, "version", bson.D{{Key: "name", Value: queryName}})
	if err != nil {
		return 0, err
	}

	queryHash, err := GetFileHash(configPath)
	if err != nil {
		return 0, err
	}

	// Document exists and we potentially need to increment the version
	if document != nil {
		var decoded VersionRecord
		err = document.Decode(&decoded)
		if err != nil {
			return 0, err
		}

		if decoded.Hash != queryHash {
			// Check if some older query version is used
			oldVersionDocument, err := mongo.GetLastDocumentFiltered(globalConfig.VersionCollectionName, "version", bson.D{{Key: "name", Value: queryName}, {Key: "hash", Value: queryHash}})
			if err != nil {
				return 0, err
			}

			if oldVersionDocument != nil {
				// Found a matching old version
				var decoded VersionRecord
				err = oldVersionDocument.Decode(&decoded)
				if err != nil {
					return 0, err
				}
				return decoded.Version, nil
			} else {
				// It is a new query which requires a version increment
				decoded.Version++
				return decoded.Version, writeQueryVersion(mongo, globalConfig, queryName, queryHash, decoded.Version)
			}
		}
		// If the hash matches then no action is required, simply return the latest version
		return decoded.Version, nil
	} else {
		// It is a new query without a version entry
		return 0, writeQueryVersion(mongo, globalConfig, queryName, queryHash, 0)
	}
}

func StartTrackers(queries []string, globalConfig Config, mongo mongodb.MongoDB, stopRequest chan any, stopResponse chan any) (err error) {
	// Reserve the "versions" name since it is used for query versioning
	for _, configPath := range queries {
		var fileName = GetFileNameWithoutExtension(configPath)
		if fileName == globalConfig.VersionCollectionName {
			return errors.New("version collection name is reserved")
		}
	}

	go func() {
		var stopChannels = []chan any{}
		for _, configPath := range queries {
			var threadStopResponse = make(chan any)
			stopChannels = append(stopChannels, threadStopResponse)
			var config = autoini.ReadIni[QueryConfig](configPath)
			config.Name = GetFileNameWithoutExtension(configPath)
			var err = mongo.CreateCollection(config.Name)
			if err != nil {
				log.Fatal(err)
			}
			config.Version, err = checkQueryVersion(mongo, globalConfig, config.Name, configPath)
			if err != nil {
				log.Fatal(err)
			}
			go trackerThread(config, mongo, stopRequest, threadStopResponse)
		}
		// Await all channels to terminate
		for _, c := range stopChannels {
			<-c
		}
		close(stopResponse)
	}()

	return
}
