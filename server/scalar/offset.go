package scalar

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
)

// Offset is a custom GraphQL type to represent an offset (large integer with possible negative values).
type Offset struct {
	v *big.Int
}

func NewOffset(v *big.Int) *Offset {
	return &Offset{v}
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (Offset) ImplementsGraphQLType(name string) bool {
	return name == "Offset"
}

// UnmarshalGraphQL is a custom unmarshaler for Offset
//
// This function will be called whenever you use the
// Offset scalar as an input.
func (t *Offset) UnmarshalGraphQL(input interface{}) error {
	v := reflect.ValueOf(input)

	//nolint:exhaustive
	switch v.Kind() {
	case reflect.String:
		v, success := new(big.Int).SetString(v.String(), 10)

		t.v = v

		if !success {
			return fmt.Errorf("conversion of %s: %w", input, ErrUnmarshall)
		}

		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t.v = new(big.Int).SetInt64(v.Int())

		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t.v = new(big.Int).SetUint64(v.Uint())

		return nil
	default:
		return fmt.Errorf("type %T: %w", input, ErrUnmarshall)
	}
}

// MarshalJSON is a custom marshaler for Offset
//
// This function will be called whenever you
// query for fields that use the Offset type.
func (t Offset) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(t.v)
	if err != nil {
		return nil, fmt.Errorf("marshalling failed: %w", err)
	}

	return bytes, nil
}

// Value returns the value for this scalar.
func (t Offset) Value() *big.Int {
	return t.v
}
