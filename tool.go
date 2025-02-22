package gokamy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DataType represents a JSON data type in the generated schema.
type DataType string

// Supported JSON data types.
const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// Definition is a struct for describing a JSON Schema.
// It includes type, description, enumeration values, properties, required fields, and additional items.
type Definition struct {
	Type                 DataType              `json:"type,omitempty"`
	Description          string                `json:"description,omitempty"`
	Enum                 []string              `json:"enum,omitempty"`
	Properties           map[string]Definition `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *Definition           `json:"items,omitempty"`
	AdditionalProperties any                   `json:"additionalProperties,omitempty"`
}

// MarshalJSON provides custom JSON marshalling for the Definition type.
// It ensures that the Properties map is initialized before marshalling.
func (d *Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*d),
	})
}

// GenerateRawSchema wraps GenerateSchema and returns the JSON marshalled schema.
// Before marshalling, it validates the generated schema using ValidateDefinition.
func GenerateRawSchema(v any) (json.RawMessage, error) {
	def, err := generateSchema(v)
	if err != nil {
		return nil, err
	}
	// Validate the generated schema internally.
	if err := ValidateDefinition(def); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}
	return json.Marshal(def)
}

// ValidateDefinition recursively validates the generated JSON Schema definition.
// It ensures that required fields exist, arrays have items defined,
// that enum values are not empty, and that if AdditionalProperties is set,
// it conforms to accepted types (bool, Definition, or *Definition).
func ValidateDefinition(def *Definition) error {
	switch def.Type {
	case Object:
		// Ensure that each required field exists in the Properties map.
		for _, req := range def.Required {
			if _, ok := def.Properties[req]; !ok {
				return fmt.Errorf("required field '%s' not defined in properties", req)
			}
		}
		// Recursively validate each property.
		for name, prop := range def.Properties {
			if err := ValidateDefinition(&prop); err != nil {
				return fmt.Errorf("invalid property '%s': %w", name, err)
			}
		}
		// Validate AdditionalProperties if set.
		if def.AdditionalProperties != nil {
			switch v := def.AdditionalProperties.(type) {
			case bool:
				// Valid – additional properties are either allowed or not.
			case Definition:
				if err := ValidateDefinition(&v); err != nil {
					return fmt.Errorf("invalid AdditionalProperties definition: %w", err)
				}
			case *Definition:
				if err := ValidateDefinition(v); err != nil {
					return fmt.Errorf("invalid AdditionalProperties definition: %w", err)
				}
			default:
				return fmt.Errorf("unsupported type for AdditionalProperties: %T", v)
			}
		}
	case Array:
		// Arrays must define the Items field.
		if def.Items == nil {
			return fmt.Errorf("array type must define 'items'")
		}
		if err := ValidateDefinition(def.Items); err != nil {
			return fmt.Errorf("invalid array items: %w", err)
		}
	case String, Number, Integer, Boolean, Null:
		// For primitive types, validate that if enum is defined, none of the values are empty.
		if len(def.Enum) > 0 {
			for i, enumVal := range def.Enum {
				if strings.TrimSpace(enumVal) == "" {
					return fmt.Errorf("enum defined but value at position %d is empty", i)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported schema type '%s'", def.Type)
	}
	return nil
}

// GenerateSchema generates a JSON schema Definition for the given value.
// It uses reflection to derive the schema based on the type of v.
func generateSchema(v any) (*Definition, error) {
	return reflectSchema(reflect.TypeOf(v))
}

// reflectSchema generates a JSON schema Definition by reflecting on the provided type.
func reflectSchema(t reflect.Type) (*Definition, error) {
	var d Definition
	switch t.Kind() {
	case reflect.String:
		d.Type = String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d.Type = Integer
	case reflect.Float32, reflect.Float64:
		d.Type = Number
	case reflect.Bool:
		d.Type = Boolean
	case reflect.Slice, reflect.Array:
		d.Type = Array
		// Recursively generate the schema for the element type.
		items, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		d.Type = Object
		// Disallow additional properties by default.
		d.AdditionalProperties = false
		objDef, err := reflectSchemaObject(t)
		if err != nil {
			return nil, err
		}
		d = *objDef
	case reflect.Ptr:
		// Dereference pointer and generate schema for the underlying type.
		definition, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d = *definition
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.UnsafePointer:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind().String())
	default:
		// Handle other unexpected types if necessary.
	}
	return &d, nil
}

// processField is a helper function that processes a struct field and generates its associated JSON schema component.
// It returns the JSON tag name, the generated schema, a flag indicating whether the field is required, and an error if any.
func processField(field reflect.StructField) (jsonTag string, schema *Definition, required bool, err error) {
	// Retrieve the JSON tag from the field.
	jsonTag = field.Tag.Get("json")
	if jsonTag == "-" {
		return "", nil, false, nil // Field is ignored.
	}
	required = true // By default, the field is required.

	if jsonTag == "" {
		jsonTag = field.Name
	} else {
		parts := strings.Split(jsonTag, ",")
		jsonTag = parts[0]
		// If 'omitempty' is specified, the field is not required.
		for _, opt := range parts[1:] {
			if strings.TrimSpace(opt) == "omitempty" {
				required = false
				break
			}
		}
	}

	// Recursively generate the schema for the field's type.
	schema, err = reflectSchema(field.Type)
	if err != nil {
		return "", nil, false, err
	}

	// Set the description if provided via the tag.
	if description := strings.TrimSpace(field.Tag.Get("description")); description != "" {
		schema.Description = description
	}

	// Handle the "enum" tag to specify enumeration values.
	if enumTag := field.Tag.Get("enum"); enumTag != "" {
		var enumValues []string
		for _, v := range strings.Split(enumTag, ",") {
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				enumValues = append(enumValues, trimmed)
			}
		}
		if len(enumValues) > 0 {
			schema.Enum = enumValues
		}
	}

	// Override the default required value using the "required" tag if provided.
	if reqTag := field.Tag.Get("required"); reqTag != "" {
		if parsed, pErr := strconv.ParseBool(reqTag); pErr == nil {
			required = parsed
		}
	}

	return jsonTag, schema, required, nil
}

// reflectSchemaObject generates a JSON schema Definition for a struct type.
// It iterates over the exported fields, processes each field, and constructs the schema properties.
func reflectSchemaObject(t reflect.Type) (*Definition, error) {
	def := Definition{
		Type:                 Object,
		AdditionalProperties: false,
	}
	properties := make(map[string]Definition)
	var requiredFields []string

	// Iterate over each field in the struct.
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		tag, schema, req, err := processField(field)
		if err != nil {
			return nil, err
		}
		// Skip fields with an empty JSON tag.
		if tag == "" {
			continue
		}

		properties[tag] = *schema
		if req {
			requiredFields = append(requiredFields, tag)
		}
	}
	def.Properties = properties
	def.Required = requiredFields
	return &def, nil
}
