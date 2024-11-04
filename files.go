package main

import (
	"log"
	"os"
	"path/filepath"
)

func ListFiles(directory string) (result []string) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
	return
}
