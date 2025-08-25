package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// go:generate ent generate ./schema

type SocialLink struct {
	ent.Schema
}

func (SocialLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("url").Unique(),
		field.Time("created_at").Default(time.Now),
		// опционально: исходный домен-источник, куда ты ходил
		field.String("source_domain").Optional(),
	}
}
