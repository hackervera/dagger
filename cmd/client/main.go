package main

import (
	"dagger"
	"log"
)

func main() {
	_, err := dagger.Client()
	if err != nil {
		log.Fatal(err)
	}
}
