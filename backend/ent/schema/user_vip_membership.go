package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserVIPMembership holds one user VIP membership period.
type UserVIPMembership struct {
	ent.Schema
}

func (UserVIPMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_vip_memberships"},
	}
}

func (UserVIPMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Int64("vip_level_id"),
		field.Time("starts_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.Int64("source_order_id").
			Optional().
			Nillable(),
		field.String("notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (UserVIPMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("vip_memberships").
			Field("user_id").
			Unique().
			Required(),
		edge.From("vip_level", VIPLevel.Type).
			Ref("memberships").
			Field("vip_level_id").
			Unique().
			Required(),
		edge.From("source_order", PaymentOrder.Type).
			Ref("vip_memberships").
			Field("source_order_id").
			Unique(),
	}
}

func (UserVIPMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("vip_level_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("source_order_id"),
		index.Fields("user_id", "status", "expires_at"),
	}
}
