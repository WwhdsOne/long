package main

import (
	"fmt"
	"log"

	"long/internaml/config"
)

func main() {
	cfg, err := config.LoadTest()
	if err != nil {
		log.Fatal(err)
	}
}
