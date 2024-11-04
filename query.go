package main

type Query struct {
	Url            string
	OverrideAnyTag string
	Before         string
	After          string
	Number         bool
	String         bool
}

func (q Query) Optional(key string) bool {
	switch key {
	case "Url":
		return false
	case "OverrideAnyTag":
		return true
	case "Before":
		return false
	case "After":
		return false
	case "Number":
		return true
	case "String":
		return true
	default:
		return false
	}
}
