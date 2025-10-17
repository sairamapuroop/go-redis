package main

import (
	"log"
	"redis-go/internal/commands"
	"redis-go/internal/db"
	"redis-go/internal/server"
)

func main() {

	// create a new in-memory database
	d := db.New()

	// create a new commands registry
	commands := commands.NewRegistry(d)

	srv := &server.Server{Address: ":6379",
		Commands: commands,
	}
	log.Println("starting kv server on port 6379")
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
