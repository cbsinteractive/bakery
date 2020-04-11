package main

import (
	"log"
	"net/http"

	"github.com/cbsinteractive/bakery/config"
	"github.com/cbsinteractive/bakery/handlers"
)

func main() {
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := c.GetLogger()
	handler := handlers.LoadHandler(c)

	logger.Infof("Starting Bakery on %s", c.Listen)
	http.Handle("/", handler)
	if err := http.ListenAndServe(c.Listen, nil); err != nil {
		log.Fatal(err)
	}
}
