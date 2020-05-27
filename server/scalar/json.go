package scalar

import (
	"encoding/json"
	"fmt"
)

// JSONObject is a custom GraphQL type to represent JSONObject.
type JSONObject struct {
	v map[string]interface{}
}

func NewJSONObject(v map[string]interface{}) *JSONObject {
	return &JSONObject{v}
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (JSONObject) ImplementsGraphQLType(name string) bool {
	return name == "JSONObject"
}

// UnmarshalGraphQL is a custom unmarshaler for JSONObject
//
// This function will be called whenever you use the
// JSONObject scalar as an input.
func (t *JSONObject) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case map[string]interface{}:
		t.v = input
		return nil
	default:
		return fmt.Errorf("type %T: %w", input, ErrUnmarshall)
	}
}

// MarshalJSON is a custom marshaler for JSONObject
//
// This function will be called whenever you
// query for fields that use the JSONObject type.
func (t JSONObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.v)
}
