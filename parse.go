package main

import (
	"bytes"
	"errors"
	"strconv"
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

func ToNumber(data string) (float64, error) {
	// TODO: make decimal point character configurable, provide more parsing options

	// Perform conversion to a number in 3 steps:
	// 1: Trim the string from the left and right by removing any non-numerical character
	// 2: Build a new string which contains only the numerical characters (ignoring any separators in the middle)
	// 3: If there are any non-numerical characters left or if there are no numerical characters at all: return an error
	var left = len(data)
	var right = -1
	var buffer bytes.Buffer
	var dotFound = false
	for i := 0; i < len(data); i++ {
		if data[i] >= '0' && data[i] <= '9' && i < left {
			left = i
			break
		}
	}
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] >= '0' && data[i] <= '9' && i > right {
			// Add 1 to be past the end of the number
			right = i + 1
			break
		}
	}

	for i := left; i < right; i++ {
		if data[i] >= '0' && data[i] <= '9' {
			buffer.WriteByte(data[i])
		} else if data[i] == '.' {
			if dotFound {
				return 0, errors.New("Invalid number in the input string: " + data)
			}
			dotFound = true
			buffer.WriteByte(data[i])
		}
	}

	return strconv.ParseFloat(buffer.String(), 64)
}
