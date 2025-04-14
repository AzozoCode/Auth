package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/azozocode/auth/initializers"
)

func seedAccount(s Storage, firstName string, lastName string, password string) *Account {
	acc, err := NewAccount(firstName, lastName, password)

	if err != nil {
		log.Fatal(err)
	}

	if err := s.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "azozo", "code", "hunter7777")
}

func init() {
	initializers.LoadEnvVariables()
}

func main() {

	seed := flag.Bool("seed", false, "seeding the db")

	flag.Parse()

	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		//seed stuff

		fmt.Println("seeding the database")
		seedAccounts(store)
	}

	server := NewAPIServer(":3000", store)
	server.Run()

}
