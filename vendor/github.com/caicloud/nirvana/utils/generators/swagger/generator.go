/*
Copyright 2018 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package swagger

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
	"github.com/go-openapi/spec"
)

var defaultSourceMapping = map[definition.Source]string{
	definition.Path:   "path",
	definition.Query:  "query",
	definition.Header: "header",
	definition.Form:   "formData",
	definition.File:   "formData",
	definition.Body:   "body",
	definition.Prefab: "",
}

var defaultDestinationMapping = map[definition.Destination]string{
	definition.Meta:  "header",
	definition.Data:  "body",
	definition.Error: "",
}

// Generator is for generating swagger specifications.
type Generator struct {
	config             *Config
	apis               *api.Definitions
	schemas            map[string]*spec.Schema
	schemaMappings     map[api.TypeName]*spec.Schema
	paths              map[string]*spec.PathItem
	sourceMapping      map[definition.Source]string
	destinationMapping map[definition.Destination]string
}

// NewDefaultGenerator creates a swagger generator with default mappings.
func NewDefaultGenerator(
	config *Config,
	apis *api.Definitions,
) *Generator {
	return NewGenerator(config, apis, nil, nil)
}

// NewGenerator creates a swagger generator.
func NewGenerator(
	config *Config,
	apis *api.Definitions,
	sourceMapping map[definition.Source]string,
	destinationMapping map[definition.Destination]string,
) *Generator {
	g := &Generator{
		config:         config,
		apis:           apis,
		schemas:        map[string]*spec.Schema{},
		schemaMappings: map[api.TypeName]*spec.Schema{},
		paths:          map[string]*spec.PathItem{},
	}
	if sourceMapping == nil {
		g.sourceMapping = defaultSourceMapping
	}
	if destinationMapping == nil {
		g.destinationMapping = defaultDestinationMapping
	}
	return g
}

// Generate generates swagger specifications.
func (g *Generator) Generate() ([]spec.Swagger, error) {
	g.parseSchemas()
	g.parsePaths()
	swaggers := []spec.Swagger{}
	for _, version := range g.config.Versions {
		title := fmt.Sprintln(g.config.Project, "APIs")
		description := g.config.Description
		if description != "" && version.Description != "" {
			description += "\n" + version.Description
		}
		schemes := version.Schemes
		if len(schemes) <= 0 {
			schemes = g.config.Schemes
		}
		hosts := version.Hosts
		if len(hosts) <= 0 {
			hosts = g.config.Hosts
		}
		contacts := version.Contacts
		if len(contacts) <= 0 {
			contacts = g.config.Contacts
		}

		swagger := g.buildSwaggerInfo(
			title, version.Name, description,
			schemes, hosts, contacts,
			version.PathRules,
		)
		swaggers = append(swaggers, *swagger)

	}
	if len(swaggers) <= 0 {
		swagger := g.buildSwaggerInfo(
			g.config.Project, "unknown", g.config.Description,
			g.config.Schemes, g.config.Hosts, g.config.Contacts,
			nil,
		)
		swaggers = append(swaggers, *swagger)
	}
	return swaggers, nil
}

func (g *Generator) buildSwaggerInfo(
	title, version, description string,
	schemes []string,
	hosts []string,
	contacts []Contact,
	rules []string,
) *spec.Swagger {
	swagger := &spec.Swagger{}
	swagger.Swagger = "2.0"
	swagger.Schemes = schemes
	if len(hosts) > 0 {
		swagger.Host = hosts[0]
	}
	swagger.Info = &spec.Info{}
	swagger.Info.Title = title
	swagger.Info.Version = version
	swagger.Info.Description = g.escapeNewline(description)
	if len(contacts) > 0 {
		swagger.Info.Contact = &spec.ContactInfo{
			Name:  contacts[0].Name,
			Email: contacts[0].Email,
		}
	}
	swagger.Definitions = spec.Definitions{}
	swagger.Paths = &spec.Paths{
		Paths: map[string]spec.PathItem{},
	}
	for path, definition := range g.schemas {
		swagger.Definitions[path] = *definition
	}
	if len(rules) > 0 {
		regexps := make([]*regexp.Regexp, 0, len(rules))
		for _, rule := range rules {
			regexps = append(regexps, regexp.MustCompile(rule))
		}
		for path, item := range g.paths {
			for _, rule := range regexps {
				if rule.FindString(path) != "" {
					swagger.Paths.Paths[path] = *item
					break
				}
			}
		}
	} else {
		for path, item := range g.paths {
			swagger.Paths.Paths[path] = *item
		}
	}
	return swagger
}

func (g *Generator) parseSchemas() {
	for _, typ := range g.apis.Types {
		g.schemaForType(typ)
	}
}

func (g *Generator) schemaForType(typ *api.Type) *spec.Schema {
	schema, ok := g.schemaMappings[typ.TypeName()]
	if !ok {
		switch typ.Kind {
		case reflect.Array, reflect.Slice:
			elem := g.schemaForTypeName(typ.Elem)
			if elem == nil {
				break
			}
			schema = spec.ArrayProperty(elem)
			schema.Title = "[]" + elem.Title
		case reflect.Ptr:
			schema = g.schemaForTypeName(typ.Elem)
		case reflect.Map:
			keySchema := g.schemaForTypeName(typ.Key)
			if keySchema == nil {
				break
			}
			elemSchema := g.schemaForTypeName(typ.Elem)
			if elemSchema == nil {
				break
			}
			schema := spec.MapProperty(elemSchema)
			schema.Title = fmt.Sprintf("map[%s]%s", keySchema.Title, elemSchema.Title)
			schema.Items = &spec.SchemaOrArray{
				Schema: keySchema,
			}
		case reflect.Struct:
			if typ.TypeName() == "time.Time" {
				schema = spec.DateTimeProperty()
				schema.Title = "Time"
			} else {
				schema = g.schemaForStruct(typ)
			}
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.String:
			schema = g.schemaForBasicType(typ)
		}
		if schema != nil {
			g.schemaMappings[typ.TypeName()] = schema
		}
	}
	return g.copySchema(schema)
}

func (g *Generator) schemaForStruct(typ *api.Type) *spec.Schema {
	const rawSchemaKey = "raw"
	typeName := typ.TypeName()
	schema, ok := g.schemaMappings[typeName]
	if ok {
		return schema
	}
	// To prevent recursive struct.
	key := strings.Replace(string(typeName), "/", "_", -1)
	ref := spec.RefSchema("#/definitions/" + key)
	ref.Title = typ.Name
	g.schemaMappings[typeName] = ref

	schema = &spec.Schema{}
	schema.Title = ref.Title
	for _, field := range typ.Fields {
		jsontag := strings.TrimSpace(field.Tag.Get("json"))
		if jsontag == "-" {
			continue
		}
		fieldSchema := g.schemaForTypeName(field.Type)
		if fieldSchema == nil {
			// Ignore invalid field.
			continue
		}
		raw, ok := fieldSchema.ExtraProps[rawSchemaKey]
		if field.Anonymous && ok {
			rawSchema := raw.(*spec.Schema)
			for name, property := range rawSchema.Properties {
				schema.SetProperty(name, property)
			}
		} else {
			name := jsontag
			if comma := strings.Index(jsontag, ","); comma > 0 {
				name = strings.TrimSpace(jsontag[:comma])
			}
			if name == "" {
				name = field.Name
			}
			fieldSchema.Description = g.escapeNewline(field.Comments)
			schema.SetProperty(name, *fieldSchema)
		}
	}
	g.schemas[key] = schema
	ref.ExtraProps = map[string]interface{}{rawSchemaKey: schema}
	return ref
}

func (g *Generator) schemaForBasicType(typ *api.Type) *spec.Schema {
	var types = map[reflect.Kind][]string{
		reflect.Bool:    {"boolean", "bool"},
		reflect.Int:     {"number", "int"},
		reflect.Int8:    {"number", "int8"},
		reflect.Int16:   {"number", "int16"},
		reflect.Int32:   {"number", "int32"},
		reflect.Int64:   {"number", "int64"},
		reflect.Uint:    {"number", "uint"},
		reflect.Uint8:   {"number", "uint8"},
		reflect.Uint16:  {"number", "uint16"},
		reflect.Uint32:  {"number", "uint32"},
		reflect.Uint64:  {"number", "uint64"},
		reflect.Uintptr: {"number", "uintptr"},
		reflect.Float32: {"number", "float32"},
		reflect.Float64: {"number", "float64"},
		reflect.String:  {"string", "string"},
	}
	formats, ok := types[typ.Kind]
	if !ok {
		return nil
	}
	schema := &spec.Schema{SchemaProps: spec.SchemaProps{
		Type: []string{formats[0]}, Format: formats[1]}}
	schema.Title = typ.Name
	return schema
}

func (g *Generator) schemaForTypeName(name api.TypeName) *spec.Schema {
	typ, ok := g.apis.Types[name]
	if !ok {
		return nil
	}
	return g.schemaForType(typ)
}

func (g *Generator) copySchema(source *spec.Schema) *spec.Schema {
	if source == nil {
		return nil
	}
	dest := *source
	return &dest
}

func (g *Generator) parsePaths() {
	for path, defs := range g.apis.Definitions {
		operations := map[string][]*spec.Operation{}
		for _, def := range defs {
			op := g.operationFor(&def)
			ops := operations[def.HTTPMethod]
			ops = append(ops, op)
			operations[def.HTTPMethod] = ops
		}
		for method, ops := range operations {
			for i, op := range ops {
				itemPath := path
				if i > 0 {
					itemPath = fmt.Sprintf("%s [%d]", path, i)
				}
				item := g.paths[itemPath]
				if item == nil {
					item = &spec.PathItem{}
					g.paths[itemPath] = item
				}
				switch method {
				case http.MethodGet:
					item.Get = op
				case http.MethodHead:
					item.Head = op
				case http.MethodPost:
					item.Post = op
				case http.MethodPut:
					item.Put = op
				case http.MethodPatch:
					item.Patch = op
				case http.MethodDelete:
					item.Delete = op
				case http.MethodOptions:
					item.Options = op
				default:
					continue
				}
			}
		}
	}
}

func (g *Generator) operationFor(def *api.Definition) *spec.Operation {
	operation := &spec.Operation{}
	consumes := map[string]bool{}
	for _, c := range def.Consumes {
		if !consumes[c] {
			consumes[c] = true
			operation.Consumes = append(operation.Consumes, c)
		}
	}
	produces := map[string]bool{}
	for _, p := range def.Produces {
		if !produces[p] {
			produces[p] = true
			operation.Produces = append(operation.Produces, p)
		}
	}
	for _, p := range def.ErrorProduces {
		if !produces[p] {
			produces[p] = true
			operation.Produces = append(operation.Produces, p)
		}
	}
	operation.Summary = def.Summary
	if operation.Summary == "" {
		// Use function name as API summary.
		typ, ok := g.apis.Types[def.Function]
		if ok {
			operation.Summary = typ.Name
		}
		if operation.Summary == "" {
			operation.Summary = "Unknown API"
		}
	}
	operation.Description = def.Description
	if operation.Description == "" {
		// Use function comments as API description.
		typ, ok := g.apis.Types[def.Function]
		if ok {
			operation.Description = typ.Comments
		}
	}
	operation.Description = g.escapeNewline(operation.Description)
	for _, param := range def.Parameters {
		parameters := g.generateParameter(&param)
		if len(parameters) > 0 {
			operation.Parameters = append(operation.Parameters, parameters...)
		}
	}
	operation.Responses = &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				def.HTTPCode: *g.generateResponse(def.Results, def.Examples),
			},
		},
	}
	return operation
}

func (g *Generator) generateParameter(param *api.Parameter) []spec.Parameter {
	if param.Source == definition.Auto {
		return g.generateAutoParameter(param.Type)
	}
	source := g.sourceMapping[param.Source]
	if source == "" {
		return nil
	}
	schema := g.schemaForTypeName(param.Type)
	parameter := spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:        param.Name,
			Description: g.escapeNewline(param.Description),
			Schema:      schema,
			In:          source,
		},
	}
	if len(param.Default) <= 0 {
		parameter.Required = true
	}
	if parameter.In != "body" {
		parameter.Type = schema.Type[0]
		parameter.Format = schema.Format
		parameter.Schema = nil
	}

	if len(param.Default) > 0 {
		r := rawJSON(param.Default)
		parameter.WithDefault(&r)
	}

	return []spec.Parameter{parameter}
}

func (g *Generator) generateAutoParameter(typ api.TypeName) []spec.Parameter {
	structType, ok := g.apis.Types[typ]
	if !ok || structType.Kind != reflect.Struct {
		return nil
	}
	return g.enum(structType)

}

func (g *Generator) enum(typ *api.Type) []spec.Parameter {
	results := []spec.Parameter{}
	for _, field := range typ.Fields {
		tag := field.Tag.Get("source")
		parameters := []spec.Parameter(nil)
		if tag != "" {
			source, name, apc, err := service.ParseAutoParameterTag(tag)
			defaultValue := apc.Get(service.AutoParameterConfigKeyDefaultValue)
			if err == nil {
				parameters = g.generateParameter(&api.Parameter{
					Source:      source,
					Name:        name,
					Description: g.escapeNewline(field.Comments),
					Type:        field.Type,
					Default:     []byte(defaultValue),
				})
			}
		} else {
			fieldType, ok := g.apis.Types[field.Type]
			if ok && fieldType.Kind == reflect.Struct {
				parameters = g.enum(fieldType)
			}
		}
		if len(parameters) > 0 {
			results = append(results, parameters...)
		}
	}
	return results
}

func (g *Generator) generateResponse(results []api.Result, examples []api.Example) *spec.Response {
	response := &spec.Response{}
	for _, result := range results {
		switch g.destinationMapping[result.Destination] {
		case "body":
			response.Description = g.escapeNewline(result.Description)
			schema := g.schemaForTypeName(result.Type)
			response.Schema = schema
		}
	}
	for _, example := range examples {
		if len(example.Instance) > 0 {
			// Only show the first example which has data.
			r := rawJSON(example.Instance)
			response.AddExample("application/json", &r)
			break
		}
	}
	return response
}
func (g *Generator) escapeNewline(content string) string {
	return strings.Replace(strings.TrimSpace(content), "\n", "<br/>", -1)
}

type rawJSON []byte

func (r *rawJSON) UnmarshalJSON(data []byte) error {
	*r = data
	return nil
}

func (r *rawJSON) MarshalJSON() ([]byte, error) {
	return *r, nil
}
