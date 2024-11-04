package main

import (
	"fmt"
	"log"

	"webtrack/autoini"
	"webtrack/webfetch"
)

func main() {
	html, err := webfetch.FetchHtml("https://en.wikipedia.org/wiki/Main_Page")

	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Println(html)
	}

	var config = autoini.ReadIni[Config]("config.ini")
	fmt.Println(config)

	var config2 = autoini.ReadIni[Query]("queries/youtube.ini")
	fmt.Println(config2)
}
