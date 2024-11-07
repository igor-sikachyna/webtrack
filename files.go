package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetFileNameWithoutExtension(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func ListIniFiles(directory string) (result []string) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Allow only .ini files
		if !info.IsDir() && strings.HasSuffix(path, ".ini") {
			result = append(result, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
	return
}
