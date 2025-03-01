package syndicate

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// --- Estructuras de prueba ---

// TestStruct se utiliza para probar reflectSchemaObject y processField.
type TestStruct struct {
	Field1 string `json:"field1" description:"Test field1" enum:"a,b,c" required:"true"`
	Field2 int    `json:"field2,omitempty" description:"Test field2" required:"false"`
	Field3 bool   `json:"-"` // Se ignora.
}

// --- Tests de processField y reflectSchemaObject ---

func TestProcessFieldAndReflectSchemaObject(t *testing.T) {
	// Obtenemos el type de TestStruct.
	typ := reflect.TypeOf(TestStruct{})
	def, err := reflectSchemaObject(typ)
	if err != nil {
		t.Fatalf("Error generating schema object: %v", err)
	}

	// Se espera que se hayan procesado Field1 y Field2.
	if len(def.Properties) != 2 {
		t.Fatalf("Se esperaban 2 propiedades, se obtuvo: %d", len(def.Properties))
	}

	// Verificar Field1
	field1, ok := def.Properties["field1"]
	if !ok {
		t.Error("Field 'field1' no encontrada en las propiedades")
	}
	if field1.Description != "Test field1" {
		t.Errorf("Descripción incorrecta en field1, se esperaba 'Test field1', se obtuvo '%s'", field1.Description)
	}
	// Enum en field1: ["a", "b", "c"]
	if len(field1.Enum) != 3 || field1.Enum[0] != "a" || field1.Enum[1] != "b" || field1.Enum[2] != "c" {
		t.Errorf("Enum incorrecto en field1, se obtuvo: %v", field1.Enum)
	}

	// Verificar que Field1 es requerida, mientras que Field2 no.
	if len(def.Required) != 1 || def.Required[0] != "field1" {
		t.Errorf("Campos requeridos incorrectos, se esperaba ['field1'] y se obtuvo: %v", def.Required)
	}
}

func TestGenerateRawSchemaSuccess(t *testing.T) {
	// Prueba de generación de schema para TestStruct.
	raw, err := GenerateRawSchema(TestStruct{})
	if err != nil {
		t.Fatalf("GenerateRawSchema falló: %v", err)
	}

	// Intentar decodificar el JSON.
	var decoded Definition
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("Error al decodificar el schema JSON: %v", err)
	}
	if decoded.Type != Object {
		t.Errorf("se esperaba 'object', se obtuvo '%s'", decoded.Type)
	}
}

func TestValidateDefinitionFailsMissingRequired(t *testing.T) {
	// Creamos un Definition con type object pero sin definir propiedad requerida.
	def := &Definition{
		Type:       Object,
		Properties: map[string]Definition{"prop1": {Type: String}},
		Required:   []string{"missingProp"},
	}
	err := ValidateDefinition(def)
	if err == nil || !strings.Contains(err.Error(), "required field 'missingProp'") {
		t.Errorf("Se esperaba error por falta de required field, se obtuvo: %v", err)
	}
}

func TestValidateDefinitionArrayItems(t *testing.T) {
	// Creamos un Definition para array sin Items definido.
	def := &Definition{
		Type: Array,
	}
	err := ValidateDefinition(def)
	if err == nil || !strings.Contains(err.Error(), "array type must define 'items'") {
		t.Errorf("Se esperaba error por array sin items, se obtuvo: %v", err)
	}
}

func TestValidateDefinitionEmptyEnumValue(t *testing.T) {
	// Para tipos primitivos, enum no debe tener valores vacíos.
	def := &Definition{
		Type: String,
		Enum: []string{"valid", "  ", "another"},
	}
	err := ValidateDefinition(def)
	if err == nil || !strings.Contains(err.Error(), "enum defined but value at position 1 is empty") {
		t.Errorf("Se esperaba error por enum con valor vacío, se obtuvo: %v", err)
	}
}

func TestAdditionalPropertiesValidation(t *testing.T) {
	// Caso válido: AdditionalProperties es bool.
	def := &Definition{
		Type:                 Object,
		Properties:           map[string]Definition{"a": {Type: String}},
		Required:             []string{"a"},
		AdditionalProperties: true,
	}
	if err := ValidateDefinition(def); err != nil {
		t.Errorf("No se esperaba error para AdditionalProperties bool, pero se obtuvo: %v", err)
	}

	// Caso válido: AdditionalProperties es Definition.
	addProp := Definition{Type: Number}
	def.AdditionalProperties = addProp
	if err := ValidateDefinition(def); err != nil {
		t.Errorf("No se esperaba error para AdditionalProperties Definition, pero se obtuvo: %v", err)
	}

	// Caso inválido: AdditionalProperties con tipo no soportado.
	def.AdditionalProperties = "unsupported"
	err := ValidateDefinition(def)
	if err == nil || !strings.Contains(err.Error(), "unsupported type for AdditionalProperties") {
		t.Errorf("Se esperaba error para AdditionalProperties invalido, se obtuvo: %v", err)
	}
}

func TestReflectSchemaPrimitiveTypes(t *testing.T) {
	// Probar tipos primitivos.
	strDef, err := reflectSchema(reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("Error en reflectSchema para string: %v", err)
	}
	if strDef.Type != String {
		t.Errorf("Se esperaba String, se obtuvo: %s", strDef.Type)
	}

	intDef, err := reflectSchema(reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("Error en reflectSchema para int: %v", err)
	}
	if intDef.Type != Integer {
		t.Errorf("Se esperaba Integer, se obtuvo: %s", intDef.Type)
	}

	boolDef, err := reflectSchema(reflect.TypeOf(true))
	if err != nil {
		t.Fatalf("Error en reflectSchema para bool: %v", err)
	}
	if boolDef.Type != Boolean {
		t.Errorf("Se esperaba Boolean, se obtuvo: %s", boolDef.Type)
	}
}

func TestReflectSchemaForArray(t *testing.T) {
	// Prueba para slices: La definición debe ser Array con Items definidos.
	var arr []string
	arrDef, err := reflectSchema(reflect.TypeOf(arr))
	if err != nil {
		t.Fatalf("Error en reflectSchema para slice: %v", err)
	}
	if arrDef.Type != Array {
		t.Errorf("Se esperaba Array, se obtuvo: %s", arrDef.Type)
	}
	if arrDef.Items == nil || arrDef.Items.Type != String {
		t.Errorf("Se esperaba que Items tenga tipo String, se obtuvo: %v", arrDef.Items)
	}
}

func TestReflectSchemaUnsupportedType(t *testing.T) {
	// Probar un tipo no soportado, p.ej. chan int.
	ch := make(chan int)
	_, err := reflectSchema(reflect.TypeOf(ch))
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("Se esperaba error para tipo no soportado (chan), se obtuvo: %v", err)
	}
}

func TestMarshalJSONDefinitionPropertiesInitialized(t *testing.T) {
	// Verificar que MarshalJSON inicialice Properties en caso de ser nil.
	def := Definition{
		Type: String,
	}
	out, err := json.Marshal(&def)
	if err != nil {
		t.Fatalf("Error en MarshalJSON: %v", err)
	}
	// El JSON resultante debe incluir "properties": {} (aunque se omita por omitempty, lo comprobamos indirectamente)
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Error al decodificar JSON: %v", err)
	}
	// Si no existe "properties", se considera que no fue inicializado en MarshalJSON.
	if _, exists := m["properties"]; !exists {
		t.Error(`Se esperaba que "properties" estuviera presente tras MarshalJSON, aunque sea un objeto vacío`)
	}
}
