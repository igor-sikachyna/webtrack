package main

import (
	"fmt"
	"log"

	"webtrack/webfetch"
)

func main() {
	html, err := webfetch.FetchHtml("https://en.wikipedia.org/wiki/Main_Page")

	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(html)
	}
}
