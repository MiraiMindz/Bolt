package bolt

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// OpenAPISpec represents OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                           `json:"openapi"`
	Info       OpenAPIInfo                      `json:"info"`
	Paths      map[string]map[string]Operation  `json:"paths"`
	Components Components                       `json:"components,omitempty"`
	Tags       []Tag                            `json:"tags,omitempty"`
}

// Tag represents an OpenAPI Tag Object for grouping operations.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// OpenAPIInfo contains API metadata
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// Operation describes a single API operation
type Operation struct {
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// Parameter describes a single operation parameter
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody describes a request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType describes media type and schema
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Response describes a single response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Schema describes data structure
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
}

// Components holds reusable schema objects
type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`
}

// GenerateDocs generates OpenAPI documentation, combining group and route docs.
func (a *App) GenerateDocs() *OpenAPISpec {
	if a.config.DocsConfig.Generator != nil {
		return a.config.DocsConfig.Generator(a)
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       a.config.DocsConfig.Title,
			Description: a.config.DocsConfig.Description,
			Version:     a.config.DocsConfig.Version,
		},
		Paths: make(map[string]map[string]Operation),
		Components: Components{
			Schemas: make(map[string]Schema),
		},
		Tags: make([]Tag, 0),
	}

	// Process groups to create top-level tags for Swagger UI
	processedGroups := make(map[string]bool)
	for _, route := range a.routes {
		if route.Group != nil && !processedGroups[route.Group.Prefix] {
			parts := strings.Split(strings.Trim(route.Group.Prefix, "/"), "/")
			tagName := parts[len(parts)-1]

			if tagName != "" {
				spec.Tags = append(spec.Tags, Tag{
					Name:        tagName,
					Description: route.Group.Doc.Summary,
				})
				processedGroups[route.Group.Prefix] = true
			}
		}
	}

	for _, route := range a.routes {
		if spec.Paths[route.Path] == nil {
			spec.Paths[route.Path] = make(map[string]Operation)
		}

		finalDoc := route.Doc
		var finalTags []string

		if route.Group != nil {
			parts := strings.Split(strings.Trim(route.Group.Prefix, "/"), "/")
			tagName := parts[len(parts)-1]
			if tagName != "" {
				finalTags = append(finalTags, tagName)
			}
			finalTags = append(finalTags, route.Group.Doc.Tags...)
			if route.Group.Doc.Description != "" {
				finalDoc.Description = route.Group.Doc.Description + "\n\n" + finalDoc.Description
			}
		}
		finalTags = append(finalTags, route.Doc.Tags...)

		operation := Operation{
			Summary:     finalDoc.Summary,
			Description: finalDoc.Description,
			Tags:        finalTags,
			Responses:   make(map[string]Response),
		}

		params := extractPathParams(route.Path)
		for _, param := range params {
			operation.Parameters = append(operation.Parameters, Parameter{
				Name:     param,
				In:       "path",
				Required: true,
				Schema:   Schema{Type: "string"},
			})
		}

		if finalDoc.Request != nil {
			schemaName := getTypeName(finalDoc.Request)
			spec.Components.Schemas[schemaName] = generateSchema(finalDoc.Request)
			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {Schema: Schema{Ref: "#/components/schemas/" + schemaName}},
				},
			}
		}

		if finalDoc.Response != nil {
			schemaName := getTypeName(finalDoc.Response)
			spec.Components.Schemas[schemaName] = generateSchema(finalDoc.Response)
			operation.Responses["200"] = Response{
				Description: "Success",
				Content: map[string]MediaType{
					"application/json": {Schema: Schema{Ref: "#/components/schemas/" + schemaName}},
				},
			}
		} else {
			operation.Responses["200"] = Response{Description: "Success"}
		}

		methodStr := strings.ToLower(string(route.Method))
		spec.Paths[route.Path][methodStr] = operation
	}

	return spec
}

func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if len(part) > 0 && (part[0] == ':' || part[0] == '*') {
			params = append(params, part[1:])
		}
	}
	return params
}

func getTypeName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	if name == "" {
		name = "Object"
	}
	return name
}

func generateSchema(v interface{}) Schema {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	schema := Schema{Type: "object", Properties: make(map[string]Schema)}
	if t.Kind() != reflect.Struct {
		return Schema{Type: getJSONType(t)}
	}
	var required []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]
		fieldSchema := Schema{Type: getJSONType(field.Type)}
		if field.Type.Kind() == reflect.Struct {
			fieldSchema = generateSchema(reflect.New(field.Type).Interface())
		}
		if field.Type.Kind() == reflect.Slice {
			itemType := field.Type.Elem()
			fieldSchema.Type = "array"
			fieldSchema.Items = &Schema{Type: getJSONType(itemType)}
			if itemType.Kind() == reflect.Struct {
				*fieldSchema.Items = generateSchema(reflect.New(itemType).Interface())
			}
		}
		schema.Properties[fieldName] = fieldSchema
		isOmitEmpty := false
		for _, part := range parts[1:] {
			if part == "omitempty" {
				isOmitEmpty = true
				break
			}
		}
		if !isOmitEmpty {
			required = append(required, fieldName)
		}
	}
	if len(required) > 0 {
		schema.Required = required
	}
	return schema
}

func getJSONType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return "string"
	}
}

// ServeSwaggerUI serves the Swagger UI HTML page.
func ServeSwaggerUI(specPath string) Handler {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            })
        }
    </script>
</body>
</html>`, specPath)

	return func(c *Context) error {
		c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response.WriteHeader(http.StatusOK)
		_, err := c.Response.Write([]byte(html))
		return err
	}
}
