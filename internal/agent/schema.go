package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ValidateJSON validates arguments against the deliberately small JSON Schema
// subset supported by AgentForge: object, string, number, integer, boolean,
// array, enum, required, properties, and additionalProperties:false.
func ValidateJSON(schema, document []byte) error {
	var schemaValue map[string]any
	schemaDecoder := json.NewDecoder(bytes.NewReader(schema))
	schemaDecoder.UseNumber()
	if err := schemaDecoder.Decode(&schemaValue); err != nil { return fmt.Errorf("invalid tool schema: %w", err) }
	var value any
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil { return fmt.Errorf("tool arguments are not valid JSON: %w", err) }
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil { return fmt.Errorf("tool arguments contain trailing JSON values") }
		return fmt.Errorf("tool arguments contain invalid trailing data: %w", err)
	}
	return validateValue("$", schemaValue, value)
}

func validateValue(path string, schema map[string]any, value any) error {
	if rawEnum, ok := schema["enum"].([]any); ok {
		matched := false
		for _, candidate := range rawEnum {
			if fmt.Sprint(candidate) == fmt.Sprint(value) && fmt.Sprintf("%T", candidate) == fmt.Sprintf("%T", value) { matched = true; break }
		}
		if !matched { return fmt.Errorf("%s must match one of the allowed enum values", path) }
	}
	kind, _ := schema["type"].(string)
	switch kind {
	case "object":
		object, ok := value.(map[string]any); if !ok { return fmt.Errorf("%s must be an object", path) }
		properties, _ := schema["properties"].(map[string]any)
		required, _ := schema["required"].([]any)
		for _, item := range required {
			key, ok := item.(string); if !ok { return fmt.Errorf("schema at %s has a non-string required entry", path) }
			if _, exists := object[key]; !exists { return fmt.Errorf("%s.%s is required", path, key) }
		}
		for key, child := range object {
			rawSchema, exists := properties[key]
			if !exists {
				if additional, configured := schema["additionalProperties"]; configured && additional == false { return fmt.Errorf("%s.%s is not allowed", path, key) }
				continue
			}
			childSchema, ok := rawSchema.(map[string]any); if !ok { return fmt.Errorf("schema for %s.%s must be an object", path, key) }
			if err := validateValue(path+"."+key, childSchema, child); err != nil { return err }
		}
	case "string":
		if _, ok := value.(string); !ok { return fmt.Errorf("%s must be a string", path) }
	case "number":
		if _, ok := value.(json.Number); !ok { return fmt.Errorf("%s must be a number", path) }
	case "integer":
		number, ok := value.(json.Number); if !ok || strings.ContainsAny(number.String(), ".eE") { return fmt.Errorf("%s must be an integer", path) }
	case "boolean":
		if _, ok := value.(bool); !ok { return fmt.Errorf("%s must be a boolean", path) }
	case "array":
		items, ok := value.([]any); if !ok { return fmt.Errorf("%s must be an array", path) }
		if itemSchema, exists := schema["items"].(map[string]any); exists {
			for index, item := range items { if err := validateValue(fmt.Sprintf("%s[%d]", path, index), itemSchema, item); err != nil { return err } }
		}
	case "":
		return fmt.Errorf("schema at %s must declare a type", path)
	default:
		return fmt.Errorf("schema at %s uses unsupported type %q", path, kind)
	}
	return nil
}
