package storage

import (
	"context"
	"log"

	"github.com/zeshi09/go_web_parser/ent"

	_ "github.com/lib/pq"
)

func Open() {
	client, err := ent.Open("postgres", "host=<host> port=<port> user=<user> dbname=<database> password=<password>")
	if err != nil {
		log.Fatalf("failed opening connection to pg: %v", err)
	}

	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
