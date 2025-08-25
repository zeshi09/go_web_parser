package schema

import "entgo.io/ent"

// SocialLink holds the schema definition for the SocialLink entity.
type SocialLink struct {
	ent.Schema
}

// Fields of the SocialLink.
func (SocialLink) Fields() []ent.Field {
	return nil
}

// Edges of the SocialLink.
func (SocialLink) Edges() []ent.Edge {
	return nil
}
