package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OpenAIDualAttempt struct {
	ent.Schema
}

func (OpenAIDualAttempt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "openai_dual_attempts"},
	}
}

func (OpenAIDualAttempt) Fields() []ent.Field {
	return []ent.Field{
		field.String("request_id").MaxLen(128).NotEmpty(),
		field.String("attempt_id").MaxLen(64).NotEmpty(),
		field.Int64("api_key_id"),
		field.Int64("user_id"),
		field.Int64("account_id").Optional().Nillable(),
		field.String("endpoint").MaxLen(128).NotEmpty(),
		field.String("method").MaxLen(16).Default("POST"),
		field.String("role").MaxLen(16).NotEmpty(),
		field.String("outcome").MaxLen(16).Default("pending"),
		field.String("service_tier").MaxLen(32).Optional().Nillable(),
		field.String("status").MaxLen(32).Default("created"),
		field.String("billing_basis").MaxLen(64).Optional().Nillable(),
		field.Float("estimated_cost").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("actual_cost").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Float("billed_cost").Default(0).SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.Time("upstream_dispatched_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("cancel_reason").MaxLen(128).Optional().Nillable(),
		field.JSON("metadata", map[string]any{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Default(time.Now).Immutable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (OpenAIDualAttempt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("request_id", "api_key_id", "attempt_id").Unique(),
		index.Fields("request_id", "api_key_id"),
		index.Fields("api_key_id", "created_at"),
		index.Fields("outcome", "created_at"),
	}
}
