package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store *PostgresStore, fname, lname, pw string, balance int64, isAdmin bool) *Account {
	acc, err := NewAccount(fname, lname, pw, balance)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.CreateAccount(acc, isAdmin); err != nil {
		log.Fatal(err)
	}

	fmt.Println("new account=>", acc.Number)

	if isAdmin {
		acc.IsAdmin = true
		if err := store.UpdateAccount(acc); err != nil {
			log.Fatal(err)
		}
	}

	return acc
}

func seedAccounts(s Storage) {
	postgresStore, ok := s.(*PostgresStore)
	if !ok {
		log.Fatal("Expected a *PostgresStore, but got a different type.")
	}

	seedAccount(postgresStore, "Goggi", "Puttar", "dhwajjain", 10, false)
}

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()
	store, err := NewPostgressStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	if *seed {
		fmt.Println("seeding the database")
		seedAccounts(store)
	}

	server := NewAPIServer(":3000", store)
	server.Run()
}

// Postgres
// docker run --name some-postgres -e POSTGRES_PASSWORD=mysecretpassword -d postgres
// pass - gobank

//Accounts and their JWT

// Admin - 46798
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo0Njc5OCwiZXhwaXJlc0F0IjoxNTAwMCwiaXNBZG1pbiI6dHJ1ZX0.QF_6k6Fm74dLNWwTYxllFE20EjdGfGZlpmlPeQVFDs8

// 440084
// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo0NDAwODQsImV4cGlyZXNBdCI6MTUwMDAsImlzQWRtaW4iOmZhbHNlfQ.XEQXlPvRxEsWdtEDfxTBxeR7hVvy90nFBVfb2mmRYXc

// {
// 	"firstname":"admin",
// 	"lastname":"2",
// 	"password":"dhwajjain",
// 	"balance":1000,
// 	"isAdmin":true
//   }
