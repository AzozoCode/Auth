package main

import (
	"log"

	"github.com/azozocode/auth/initializers"
)

func init() {
	initializers.LoadEnvVariables()
}

func main() {

	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", store)
	server.Run()

}
