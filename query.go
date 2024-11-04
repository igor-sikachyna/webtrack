package main

import "log"

type QueryConfig struct {
	Url                    string
	AnyTag                 string
	Before                 string
	After                  string
	ResultType             string
	RequestBackend         string
	RequestIntervalSeconds int
}

func (q QueryConfig) Optional(key string) bool {
	switch key {
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

func (q *QueryConfig) PostInit() {
	if q.ResultType != "string" && q.ResultType != "number" {
		log.Fatalln("Only \"string\" and \"number\" result types are supported")
	}
}
