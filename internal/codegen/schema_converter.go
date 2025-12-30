package codegen

import (
	"fmt"
	"strings"

	"github.com/yousuf/codebraid-mcp/internal/strutil"
)

// SchemaConverter converts JSON Schema to TypeScript types
type SchemaConverter struct {
	generatedTypes map[string]*TSType // Track generated types to avoid duplicates
}

// NewSchemaConverter creates a new schema converter
func NewSchemaConverter() *SchemaConverter {
	return &SchemaConverter{
		generatedTypes: make(map[string]*TSType),
	}
}

// ConvertSchema converts a JSON Schema to a TypeScript type
func (sc *SchemaConverter) ConvertSchema(schema map[string]interface{}, typeName string) (*TSType, error) {
	if schema == nil {
		return &TSType{
			Kind:    "primitive",
			RawType: "any",
		}, nil
	}

	// Check if already generated
	if existing, ok := sc.generatedTypes[typeName]; ok {
		return existing, nil
	}

	tsType := &TSType{
		Name: typeName,
	}

	// Get description if available
	if desc, ok := schema["description"].(string); ok {
		tsType.Description = desc
	}

	// Handle type
	schemaType, hasType := schema["type"]
	if !hasType {
		// Check for oneOf, anyOf, allOf
		if oneOf, ok := schema["oneOf"].([]interface{}); ok {
			return sc.convertUnion(oneOf, typeName)
		}
		if anyOf, ok := schema["anyOf"].([]interface{}); ok {
			return sc.convertUnion(anyOf, typeName)
		}
		if allOf, ok := schema["allOf"].([]interface{}); ok {
			return sc.convertIntersection(allOf, typeName)
		}

		// Default to any
		return &TSType{
			Kind:    "primitive",
			RawType: "any",
		}, nil
	}

	// Handle type as string or array of strings
	switch t := schemaType.(type) {
	case string:
		return sc.convertSingleType(schema, t, typeName)
	case []interface{}:
		// Union type like ["string", "null"]
		return sc.convertTypeArray(t, typeName)
	default:
		return nil, fmt.Errorf("invalid type format: %T", schemaType)
	}
}

// convertSingleType handles a single type string
func (sc *SchemaConverter) convertSingleType(schema map[string]interface{}, typeStr string, typeName string) (*TSType, error) {
	switch typeStr {
	case "string":
		// Check for enum
		if enum, ok := schema["enum"].([]interface{}); ok {
			return sc.convertEnum(enum, typeName)
		}
		return &TSType{
			Kind:    "primitive",
			RawType: "string",
		}, nil

	case "number", "integer":
		return &TSType{
			Kind:    "primitive",
			RawType: "number",
		}, nil

	case "boolean":
		return &TSType{
			Kind:    "primitive",
			RawType: "boolean",
		}, nil

	case "null":
		return &TSType{
			Kind:    "primitive",
			RawType: "null",
		}, nil

	case "array":
		return sc.convertArray(schema, typeName)

	case "object":
		return sc.convertObject(schema, typeName)

	default:
		return &TSType{
			Kind:    "primitive",
			RawType: "any",
		}, nil
	}
}

// convertTypeArray handles type as array (union)
func (sc *SchemaConverter) convertTypeArray(types []interface{}, typeName string) (*TSType, error) {
	unionTypes := make([]*TSType, 0, len(types))

	for i, t := range types {
		typeStr, ok := t.(string)
		if !ok {
			continue
		}

		subType, err := sc.convertSingleType(nil, typeStr, fmt.Sprintf("%s_%d", typeName, i))
		if err != nil {
			return nil, err
		}
		unionTypes = append(unionTypes, subType)
	}

	if len(unionTypes) == 0 {
		return &TSType{
			Kind:    "primitive",
			RawType: "any",
		}, nil
	}

	if len(unionTypes) == 1 {
		return unionTypes[0], nil
	}

	return &TSType{
		Kind:       "union",
		Name:       typeName,
		UnionTypes: unionTypes,
	}, nil
}

// convertObject converts an object schema to TypeScript interface
func (sc *SchemaConverter) convertObject(schema map[string]interface{}, typeName string) (*TSType, error) {
	properties, hasProperties := schema["properties"].(map[string]interface{})
	additionalProps, hasAdditionalProps := schema["additionalProperties"]

	// Handle Record<string, T> pattern
	if !hasProperties && hasAdditionalProps {
		if additionalPropsSchema, ok := additionalProps.(map[string]interface{}); ok {
			valueType, err := sc.ConvertSchema(additionalPropsSchema, typeName+"Value")
			if err != nil {
				return nil, err
			}
			return &TSType{
				Kind:    "type",
				Name:    typeName,
				RawType: fmt.Sprintf("Record<string, %s>", sc.typeToString(valueType)),
			}, nil
		} else if additionalProps == true {
			return &TSType{
				Kind:    "type",
				Name:    typeName,
				RawType: "Record<string, any>",
			}, nil
		}
	}

	// Regular object with properties
	if !hasProperties {
		return &TSType{
			Kind:    "type",
			Name:    typeName,
			RawType: "Record<string, any>",
		}, nil
	}

	required := make(map[string]bool)
	if reqArray, ok := schema["required"].([]interface{}); ok {
		for _, r := range reqArray {
			if reqStr, ok := r.(string); ok {
				required[reqStr] = true
			}
		}
	}

	tsProperties := make([]TSProperty, 0, len(properties))

	for propName, propSchema := range properties {
		propSchemaMap, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		propTypeName := typeName + strutil.ToPascalCase(propName)
		propType, err := sc.ConvertSchema(propSchemaMap, propTypeName)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property %q: %w", propName, err)
		}

		prop := TSProperty{
			Name:       propName,
			Type:       propType,
			IsOptional: !required[propName],
		}

		if desc, ok := propSchemaMap["description"].(string); ok {
			prop.Description = desc
		}

		tsProperties = append(tsProperties, prop)
	}

	tsType := &TSType{
		Kind:       "interface",
		Name:       typeName,
		Properties: tsProperties,
	}

	if desc, ok := schema["description"].(string); ok {
		tsType.Description = desc
	}

	sc.generatedTypes[typeName] = tsType
	return tsType, nil
}

// convertArray converts an array schema
func (sc *SchemaConverter) convertArray(schema map[string]interface{}, typeName string) (*TSType, error) {
	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		return &TSType{
			Kind: "array",
			ElementType: &TSType{
				Kind:    "primitive",
				RawType: "any",
			},
		}, nil
	}

	elementType, err := sc.ConvertSchema(items, typeName+"Item")
	if err != nil {
		return nil, err
	}

	return &TSType{
		Kind:        "array",
		Name:        typeName,
		ElementType: elementType,
	}, nil
}

// convertEnum converts an enum to a union of literals
func (sc *SchemaConverter) convertEnum(enumValues []interface{}, typeName string) (*TSType, error) {
	unionTypes := make([]*TSType, 0, len(enumValues))

	for _, val := range enumValues {
		var rawType string
		switch v := val.(type) {
		case string:
			rawType = fmt.Sprintf("%q", v)
		case float64:
			rawType = fmt.Sprintf("%v", v)
		case bool:
			rawType = fmt.Sprintf("%v", v)
		default:
			rawType = fmt.Sprintf("%v", v)
		}

		unionTypes = append(unionTypes, &TSType{
			Kind:    "primitive",
			RawType: rawType,
		})
	}

	return &TSType{
		Kind:       "union",
		Name:       typeName,
		UnionTypes: unionTypes,
	}, nil
}

// convertUnion converts oneOf/anyOf to union type
func (sc *SchemaConverter) convertUnion(schemas []interface{}, typeName string) (*TSType, error) {
	unionTypes := make([]*TSType, 0, len(schemas))

	for i, schema := range schemas {
		schemaMap, ok := schema.(map[string]interface{})
		if !ok {
			continue
		}

		subTypeName := fmt.Sprintf("%s_%d", typeName, i)
		subType, err := sc.ConvertSchema(schemaMap, subTypeName)
		if err != nil {
			return nil, err
		}

		unionTypes = append(unionTypes, subType)
	}

	if len(unionTypes) == 0 {
		return &TSType{
			Kind:    "primitive",
			RawType: "any",
		}, nil
	}

	if len(unionTypes) == 1 {
		return unionTypes[0], nil
	}

	return &TSType{
		Kind:       "union",
		Name:       typeName,
		UnionTypes: unionTypes,
	}, nil
}

// convertIntersection converts allOf to intersection type
func (sc *SchemaConverter) convertIntersection(schemas []interface{}, typeName string) (*TSType, error) {
	// For now, merge all properties into a single interface
	// A more sophisticated approach would use TypeScript intersection types

	allProperties := make([]TSProperty, 0)
	description := ""

	for i, schema := range schemas {
		schemaMap, ok := schema.(map[string]interface{})
		if !ok {
			continue
		}

		subTypeName := fmt.Sprintf("%s_%d", typeName, i)
		subType, err := sc.ConvertSchema(schemaMap, subTypeName)
		if err != nil {
			return nil, err
		}

		if subType.Kind == "interface" {
			allProperties = append(allProperties, subType.Properties...)
			if description == "" && subType.Description != "" {
				description = subType.Description
			}
		}
	}

	return &TSType{
		Kind:        "interface",
		Name:        typeName,
		Properties:  allProperties,
		Description: description,
	}, nil
}

// typeToString converts a TSType to its string representation
func (sc *SchemaConverter) typeToString(t *TSType) string {
	if t == nil {
		return "any"
	}

	switch t.Kind {
	case "primitive":
		return t.RawType
	case "array":
		if t.ElementType != nil {
			return sc.typeToString(t.ElementType) + "[]"
		}
		return "any[]"
	case "union":
		parts := make([]string, len(t.UnionTypes))
		for i, ut := range t.UnionTypes {
			parts[i] = sc.typeToString(ut)
		}
		return strings.Join(parts, " | ")
	case "interface", "type":
		if t.Name != "" {
			return t.Name
		}
		return "any"
	default:
		return "any"
	}
}
