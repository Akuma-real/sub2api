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

// VIPLevel holds the schema definition for sellable VIP membership levels.
type VIPLevel struct {
	ent.Schema
}

func (VIPLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "vip_levels"},
	}
}

func (VIPLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("original_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Optional().
			Nillable(),
		field.Int("validity_days").
			Default(30),
		field.Float("discount_multiplier").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Default(1),
		field.String("features").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.JSON("benefits", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Bool("for_sale").
			Default(true),
		field.Int("sort_order").
			Default(0),
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

func (VIPLevel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("memberships", UserVIPMembership.Type),
		edge.To("usage_logs", UsageLog.Type),
	}
}

func (VIPLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("for_sale"),
		index.Fields("sort_order"),
	}
}
