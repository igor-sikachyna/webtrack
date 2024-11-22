package main

import (
	"errors"
)

type QueryConfig struct {
	Name                   string // Internal only
	Version                int64  // Internal only
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
	case "Version":
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

func (q *QueryConfig) PostInit() (err error) {
	if q.ResultType != "string" && q.ResultType != "number" {
		return errors.New("Invalid result type " + q.ResultType + ". Only \"string\" and \"number\" result types are supported")
	}
	if q.RequestBackend != "chrome" && q.RequestBackend != "go" {
		return errors.New("Invalid request backend " + q.RequestBackend + ". Only \"chrome\" and \"go\" request backends are supported")
	}
	return
}
