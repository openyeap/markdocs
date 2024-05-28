package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {

	conf := flag.String("c", "site.json", "config file")
	flag.Parse()
	site := NewSite(*conf)

	site.Public()
	go site.Server()
	http.Handle("/", &site)
	err := http.ListenAndServe(":5555", nil)
	if nil != err {
		log.Fatalln("ERROR:", err.Error())
	}
}
