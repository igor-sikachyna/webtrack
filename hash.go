package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func GetFileHash(path string) (hash string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	var hashFunction = sha256.New()
	_, err = io.Copy(hashFunction, file)
	if err != nil {
		return
	}

	hash = fmt.Sprintf("%x", hashFunction.Sum(nil))
	return
}
