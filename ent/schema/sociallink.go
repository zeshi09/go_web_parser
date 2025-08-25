package schema

import (
	// "time"

	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SocialLink holds the schema definition for the SocialLink entity.
type SocialLink struct {
	ent.Schema
}

// Fields of the SocialLink.
func (SocialLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("link").
			Comment("Social media link/identifier"),
		field.String("url").
			Comment("Full social media URL"),
		field.String("domain").
			Optional().
			Comment("Social media domain (t.me, vk.com, etc.)"),
		field.String("source_domain").
			Optional().
			Comment("Domain where this link was found"),
		field.Time("created_at").
			Default(time.Now).
			Comment("When this link was discovered"),
	}
}

// Edges of the SocialLink.
func (SocialLink) Edges() []ent.Edge {
	return nil
}

// Indexes of the SocialLink.
func (SocialLink) Indexes() []ent.Index {
	return []ent.Index{
		// Уникальный индекс по URL чтобы избежать дубликатов
		index.Fields("url").Unique(),
		// Индекс по домену для быстрого поиска
		index.Fields("domain"),
		// Индекс по источнику
		index.Fields("source_domain"),
		// Составной индекс
		index.Fields("domain", "created_at"),
	}
}
