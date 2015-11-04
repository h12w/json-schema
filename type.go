package schema

import (
	"encoding/json"
)

type Schema struct {
	ID                   string                 `json:"id,omitempty"`
	Schema               *Ref                   `json:"$schema,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Default              interface{}            `json:"default,omitempty"`
	MultipleOf           float64                `json:"multipleOf,omitempty"`
	Maximum              float64                `json:"maximum,omitempty"`
	ExclusiveMaximum     bool                   `json:"exclusiveMaximum,omitempty"`
	Minimum              float64                `json:"minimum,omitempty"`
	ExclusiveMinimum     bool                   `json:"exclusiveMinimum,omitempty"`
	MaxLength            int                    `json:"maxLength,omitempty"`
	MinLength            int                    `json:"minLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	AdditionalItems      interface{}            `json:"additionalItems,omitempty"`
	Items                *Schema                `json:"items,omitempty"`
	MaxItems             int                    `json:"maxItems,omitempty"`
	MinItems             int                    `json:"minItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty"`
	MaxProperties        int                    `json:"maxProperties"`
	MinProperties        int                    `json:"minProperties"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties *Schema                `json:"additionalProperties,omitempty"`
	Definitions          map[string]*Schema     `json:"definitions,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty"`
	PatternProperties    map[string]*Schema     `json:"patternProperties,omitempty"`
	Dependencies         map[string]interface{} `json:"dependencies,omitempty"`
	Enum                 []string               `json:"enum,omitempty"`
	AllOf                []*Schema              `json:"allOf,omitempty"`
	AnyOf                []*Schema              `json:"anyOf,omitempty"`
	OneOf                []*Schema              `json:"oneOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty"`
	Type                 interface{}            `json:"type,omitempty"`

	// Meta-only
	Ref    *Ref   `json:"$ref,omitempty"`
	Format string `json:"format,omitempty"`
}

type Ref string

func (s *Schema) String() string {
	buf, _ := json.MarshalIndent(s, "", "\t")
	return string(buf)
}
