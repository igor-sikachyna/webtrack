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
	var allValues = map[string]struct{}{}
	if config.OnlyIfDifferent {
		var lastDocumentBson, err = mongo.GetLastDocument(config.Name, "timestamp")
		if lastDocumentBson != nil {
			if err != nil {
				log.Fatal(err)
			}

			lastDocument, err := mongodb.BsonToRaw(lastDocumentBson)
			if err != nil {
				log.Fatal(err)
			}
			lastValue = lastDocument.Lookup("value").StringValue()
		}
	}

	// TODO: Use MongoDB to check for uniqueness
	if config.OnlyIfUnique {
		var documentBsons, err = mongo.GetAllDocuments(config.Name)
		if err != nil {
			log.Fatal(err)
		}
		for _, documentBson := range documentBsons {
			document, err := mongodb.BsonToRaw(documentBson)
			if err != nil {
				log.Fatal(err)
			}
			var value = document.Lookup("value").StringValue()
			allValues[value] = struct{}{}
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
					var onlyIfUniquePassed = true
					if config.OnlyIfUnique {
						_, seen := allValues[res]
						onlyIfUniquePassed = !seen
					}

					if err == nil && onlyIfDifferentPassed && onlyIfUniquePassed {
						var timestamp = time.Now().Unix()
						err = mongo.Write(config.Name, bson.D{{Key: "timestamp", Value: timestamp}, {Key: "value", Value: res}, {Key: "version", Value: config.Version}})
						if err != nil {
							fmt.Printf("Failed to write to MongoDB: %v", err)
						} else {
							fmt.Printf("Wrote to MongoDB collection %v at %v\n", config.Name, timestamp)
							lastValue = res

							if config.OnlyIfUnique {
								allValues[res] = struct{}{}
							}
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
	return mongo.Write(globalConfig.VersionCollectionName, bson.D{{Key: "version", Value: version}, {Key: "name", Value: queryName}, {Key: "hash", Value: queryHash}})
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
	fmt.Println(queryName)
	fmt.Println(queryHash)

	// Document exists and we potentially need to increment the version
	if len(document) > 0 {
		rawDocument, err := mongodb.BsonToRaw(document)
		if err != nil {
			return 0, err
		}
		var hash = rawDocument.Lookup("hash").StringValue()
		var version, ok = rawDocument.Lookup("version").AsInt64OK()
		if !ok {
			return 0, errors.New("failed to read version number from MongoDB")
		}
		if hash != queryHash {
			// Check if some older query version is used
			oldVersionDocument, err := mongo.GetLastDocumentFiltered(globalConfig.VersionCollectionName, "version", bson.D{{Key: "name", Value: queryName}, {Key: "hash", Value: queryHash}})
			if err != nil {
				return 0, err
			}

			if len(oldVersionDocument) > 0 {
				// Found a matching old version
				rawDocument, err := mongodb.BsonToRaw(oldVersionDocument)
				if err != nil {
					return 0, err
				}
				version, ok = rawDocument.Lookup("version").AsInt64OK()
				if !ok {
					return 0, errors.New("failed to read version number from MongoDB")
				}
				return version, nil
			} else {
				// It is a new query which requires a version increment
				version++
				return version, writeQueryVersion(mongo, globalConfig, queryName, queryHash, version)
			}
		}
		// If the hash matches then no action is required, simply return the latest version
		return version, nil
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
