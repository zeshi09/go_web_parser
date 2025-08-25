package storage

import (
	"context"
	"entgo.io/ent/dialect"
	"github.com/zeshi09/go_web_parser/ent"
)

func Open(dsn string) (*ent.Client, func(), error) {
	driver := ent.Driver(dialect.Postgres)

	client, err := ent.Open(driver, dsn)
	if err != nil {
		return nil, nil, err
	}


	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		return nil, nil, err
	}

	cleanup := func() { _ = client.Close() }
	
	return client, cleanup, nil
}
