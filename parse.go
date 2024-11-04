package main

import (
	"errors"
	"strings"
)

func findIndex(data string, parts []string, moveIndexToTheEnd bool) (idx int) {
	idx = 0
	for _, part := range parts {
		idx = idx + strings.Index(data[idx:], part)
		if idx < 0 {
			return
		}
	}
	if moveIndexToTheEnd {
		return idx + len(parts[len(parts)-1])
	}
	return idx
}

func ExtractValueFromString(data string, before string, after string, anyTag string) (string, error) {
	var beforeArray = strings.Split(before, anyTag)
	var afterArray = strings.Split(after, anyTag)
	var begin = findIndex(data, beforeArray, true)
	if begin < 0 {
		return "", errors.New("beginning not found")
	}
	var end = findIndex(data[begin:], afterArray, false)
	if end < 0 {
		return "", errors.New("ending not found")
	}
	return data[begin:(begin + end)], nil
}
