package main

import (
	"log"

	"github.com/jdgonzalez907/1channel/internal/config"
)

func main() {
	cfg, err := config.NewConfiguration()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%+v\n", cfg)
	}
}
