package main

import "log"

type QueryConfig struct {
	Name                   string
	Url                    string
	AnyTag                 string
	Before                 string
	After                  string
	ResultType             string
	RequestBackend         string
	RequestIntervalSeconds int
	OnlyIfDifferent        bool
	OnlyIfUnique           bool
}

func (q QueryConfig) Optional(key string) bool {
	switch key {
	case "Name":
		return true
	case "Url":
		return false
	case "AnyTag":
		return true
	case "Before":
		return false
	case "After":
		return false
	case "ResultType":
		return true
	case "RequestBackend":
		return true
	case "RequestIntervalSeconds":
		return true
	case "OnlyIfDifferent":
		return true
	case "OnlyIfUnique":
		return true
	default:
		return false
	}
}

func (q QueryConfig) DefaultString(key string) string {
	switch key {
	case "AnyTag":
		return "<any>"
	case "ResultType":
		return "string"
	case "RequestBackend":
		return "go"
	default:
		return ""
	}
}

func (q QueryConfig) DefaultInt(key string) int {
	switch key {
	case "RequestIntervalSeconds":
		return 1
	default:
		return 0
	}
}

func (q QueryConfig) DefaultBool(key string) bool {
	return false
}

func (q *QueryConfig) PostInit() {
	if q.ResultType != "string" && q.ResultType != "number" {
		log.Fatalf("Invalid result type %v. Only \"string\" and \"number\" result types are supported", q.ResultType)
	}
	if q.RequestBackend != "chrome" && q.RequestBackend != "go" {
		log.Fatalf("Invalid request backend %v. Only \"chrome\" and \"go\" request backends are supported", q.RequestBackend)
	}
}
