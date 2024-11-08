# webtrack - Real-time webpage value tracker in Go

![tests](https://github.com/igor-sikachyna/webtrack/actions/workflows/run-tests.yml/badge.svg) ![build](https://github.com/igor-sikachyna/webtrack/actions/workflows/create-release.yml/badge.svg)

Application written in Go to collect the information from webpages or API endpoints and to store it in MongoDB

## Prerequisites

- Tested with Go 1.23.2
- [MongoDB](https://www.mongodb.com/docs/manual/administration/install-community/)

## Installation

To install and run the application run the following set of commands:

```sh
git clone git@github.com:igor-sikachyna/webtrack.git
cd webtrack
go install
go run .
```

This will clone the repo, install the dependencies, and run the application.

## How to use

To customize the behavior of the application, you need to edit the `config.ini` files with the desired values and create the query files in `queries` directory (by default).

The available options for those configurations are provided in the sections below.

Each query file created will be used to track a single value in the configured website or API endpoint and will store the collected values in the MongoDB collection with the same name as the query file.

Note that at the moment you cannot customize the collection name, the file name will be used to determine the collection name.

## Global configuration settings

| Parameter             | Is optional | Default value | Description                                                                                                                                                                                       |
| --------------------- | ----------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| MongodbConnectionUrl  | No          | N/A           | MongoDB server connection URL. Can be either left unchanged from the example config (if using a local installation) or updated to the connection URL you want to use (e.g. for a remote database) |
| DatabaseName          | No          | N/A           | Database name to create and use in MongoDB. If the database already exists, no action is performed. This requires a permission to create databases                                                |
| VersionCollectionName | No          | N/A           | Collection name to use for query versioning information                                                                                                                                           |

## Query configuration

### How to determine the `Before` and `After` values

| Parameter              | Is optional | Default value | Description                                                                                                                                                                                                                                                                                                                                                   |
| ---------------------- | ----------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Url                    | No          | N/A           | URL to query. Can include URL-encoded arguments                                                                                                                                                                                                                                                                                                               |
| AnyTag                 | Yes         | `<any>`       | String to be used as a wildcard. When processing the HTML file, `webtrack` will treat this value the same way the `*` is treated as a wildcard on Linux                                                                                                                                                                                                       |
| Before                 | No          | N/A           | String value to search for. Everything that is past the `Before` and in front of `After` will be used to determine the final fetched value. You can put the `<any>` tag (or the overriding value in `AnyTag`) to skip variable sections of the HTML file.                                                                                                     |
| After                  | No          | N/A           | String value to search for. Everything that is past the `Before` and in front of `After` will be used to determine the final fetched value. You can put the `<any>` tag (or the overriding value in `AnyTag`) to skip variable sections of the HTML file.                                                                                                     |
| ResultType             | Yes         | `string`      | Can be either `number` or `string`. The `string` type will result in the full string between `Before` and `After` to be stored in MongoDB. The `number` type will attempt to extract a single number from the resulting string; it is up to you to ensure that the value enclosed between `Before` and `After` can reasonably be converted to a single number |
| RequestBackend         | Yes         | `go`          | Determines the flow that will be used to fetch the HTML page. The `go` backend will rely on the standard Go HTTP package. The `chrome` backend will use the Chrome browser to load the HTML content. See the note below for details on `chrome` option                                                                                                        |
| RequestIntervalSeconds | Yes         | 1             | Interval in seconds between requests. This interval includes the time it takes to perform the request itself. If the request takes longer than `RequestIntervalSeconds`, then the next request will happen right after the previous one                                                                                                                       |
| OnlyIfDifferent        | Yes         | `false`       | Setting this option to `true` will make it so the values are written to MongoDB only if they changed since the last request was made                                                                                                                                                                                                                          |
| OnlyIfUnique           | Yes         | `false`       | Setting this option to `true` will make it so the values are written to MongoDB only if they don't already exist in this collection                                                                                                                                                                                                                           |

### Note about the `RequestBackend` parameter

For some websites, the standard Go HTTP request package will not be able to fully load the page, as it may require JavaScript to load the content.

For such scenarios, the `RequestBackend` should be set to `chrome`. This will use Chrome browser to load the page which includes the JavaScript content required to make certain websites load properly.

Please note that this configuration requires Chrome browser to be installed on the system and to be available in `PATH`. Additionally, the CPU and RAM usage will increase drastically due to the nature of using a fully-fledged browser to perform such requests. The request time will also increase.

## Example queries

Some queries are already provided in this repository to demonstrate the functionality:

- `stackoverflow.ini` - collects the newest unique post titles from StackOverflow
- `weather.ini` - tracks the current temperature in New York based on The Weather Network data
- `wikipedia.ini` - tracks the current total number of articles in English Wikipedia
- `youtube.ini` - tracks the current number of views of the Crab Rave music video on YouTube


## Query versioning

The `_versions` collection is used to track the changes in queries that you perform. Each modification of the query `.ini` file will produce a new entry in the `_versions` collection.

The new version number will be included on all entries entered into the MongoDB collection for the specific query.