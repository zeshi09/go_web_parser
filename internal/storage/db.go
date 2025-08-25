package storage

import (
	"context"
	"fmt"
	"entgo.io/ent/dialect"
	"github.com/zeshi09/go_web_parser/ent"
)

func Open(dsn string) (*ent.Client, func(), error) {
	driver := ent.Driver(dialect.Postgres)

	client, err := ent.Open(driver, dsn)
	if err != nil {
		return nil, nil, err
	}


	
	return client, cleanup, nil
}
