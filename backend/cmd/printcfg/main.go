package main

import (
	"fmt"
	"log"

	"long/internal/config"
)

func main() {
	cfg, err := config.LoadTest()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
