package gokamy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type Tool interface {
	GetDefinition() *openai.FunctionDefinition
	Execute(args json.RawMessage) (any, error)
}

type DataType string

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
// It is fairly limited, and you may have better luck using a third-party library.
type Definition struct {
	Type                 DataType              `json:"type,omitempty"`
	Description          string                `json:"description,omitempty"`
	Enum                 []string              `json:"enum,omitempty"`
	Properties           map[string]Definition `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *Definition           `json:"items,omitempty"`
	AdditionalProperties any                   `json:"additionalProperties,omitempty"`
}

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

func GenerateSchema(v any) (*Definition, error) {
	return reflectSchema(reflect.TypeOf(v))
}

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
		items, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		d.Type = Object
		d.AdditionalProperties = false
		objDef, err := reflectSchemaObject(t)
		if err != nil {
			return nil, err
		}
		d = *objDef
	case reflect.Ptr:
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
		// Se puede definir un caso default para tipos inesperados si se requiere.
	}
	return &d, nil
}

func reflectSchemaObject(t reflect.Type) (*Definition, error) {
	def := Definition{
		Type:                 Object,
		AdditionalProperties: false,
	}

	properties := make(map[string]Definition)
	var requiredFields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Determina el nombre de la propiedad con base en el tag "json".
		jsonTag := field.Tag.Get("json")
		// Si el tag es "-", se omite el campo.
		if jsonTag == "-" {
			continue
		}
		var required = true
		if jsonTag == "" {
			jsonTag = field.Name
		} else {
			// Si el tag incluye opciones, se extrae el primero (nombre).
			parts := strings.Split(jsonTag, ",")
			jsonTag = parts[0]
			// Se marca como no requerido si se especifica "omitempty".
			for _, opt := range parts[1:] {
				if strings.TrimSpace(opt) == "omitempty" {
					required = false
					break
				}
			}
		}

		// Genera el schema recursivamente para el tipo del campo.
		item, err := reflectSchema(field.Type)
		if err != nil {
			return nil, err
		}

		// Asigna la descripción, si está presente.
		if description := strings.TrimSpace(field.Tag.Get("description")); description != "" {
			item.Description = description
		}

		// Manejo robusto del tag "enum".
		if enumTag := field.Tag.Get("enum"); enumTag != "" {
			var enumValues []string
			// Separa y limpia los posibles valores.
			for _, v := range strings.Split(enumTag, ",") {
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					enumValues = append(enumValues, trimmed)
				}
			}
			if len(enumValues) > 0 {
				item.Enum = enumValues
			}
		}

		properties[jsonTag] = *item

		// Revisa si se definió el tag "required" para sobreescribir.
		if reqTag := field.Tag.Get("required"); reqTag != "" {
			if parsed, err := strconv.ParseBool(reqTag); err == nil {
				required = parsed
			}
		}
		if required {
			requiredFields = append(requiredFields, jsonTag)
		}
	}
	def.Required = requiredFields
	def.Properties = properties
	return &def, nil
}
