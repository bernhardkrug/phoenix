package main

import (
	"github.com/bernhardkrug/phoenix"
	"log"
)

func main() {
	db, err := connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	sqlInterface, err := db.DB()
	if err != nil {
		log.Fatal(err)

	}

	phoenix.Rise(sqlInterface, phoenix.Postgres, phoenix.WithSchema("ddhub"))
}
