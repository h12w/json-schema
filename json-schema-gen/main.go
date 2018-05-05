package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"h12.io/gengo"
	"h12.io/json-schema"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("json-schema-gen [name map file] [filenames...]")
		return
	}
	nameMap, err := readNameMap(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	g := generator{nameMap}
	var decls []*gengo.TypeDecl
	for _, filename := range os.Args[2:] {
		ds, err := g.collectDecls(filename)
		if err != nil {
			log.Fatal(err)
		}
		decls = append(decls, ds...)
	}
	codeFile := gengo.File{
		PackageName: "openrtb",
		TypeDecls:   g.filterDecls(decls),
		Imports: []*gengo.Import{
			{Path: "h12.io/decimal"},
		},
	}
	sort.Sort(codeFile.TypeDecls)
	if err := codeFile.Marshal(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

type generator struct {
	nameMap
}

func (g *generator) collectDecls(filename string) ([]*gengo.TypeDecl, error) {
	var s schema.Schema
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	decls, err := g.goTypeDecls(s.ID, &s)
	if err != nil {
		return nil, err
	}
	return decls, nil
}

func (g *generator) filterDecls(decls []*gengo.TypeDecl) (res []*gengo.TypeDecl) {
	m := make(map[string]*gengo.TypeDecl)
	for _, decl := range decls {
		if m[decl.Name] == nil {
			m[decl.Name] = decl
		}
	}

	for _, decl := range m {
		for _, field := range decl.Type.Fields {
			if field.Type.Kind == gengo.IdentKind {
				if t, ok := m[field.Type.Ident]; ok && t.Type.Kind != gengo.IdentKind || !ok && !isSimpleType(field.Type.Ident) {
					field.Type.Ident = "*" + field.Type.Ident
				}
			}
		}
		switch decl.Name {
		case "PositiveInt", "BooleanInt":
		default:
			res = append(res, decl)
		}
	}
	return
}

func isSimpleType(s string) bool {
	switch s {
	case "string", "int", "bool", "BoolInt", "float32", "float64", "interface{}", "decimal.D":
		return true
	}
	return false
}

func (g *generator) goTypeDecls(id string, s *schema.Schema) ([]*gengo.TypeDecl, error) {
	if len(s.Properties) == 0 {
		if typ, ok := s.Type.(string); ok {
			if identType, err := g.goIdentType(id, typ); err == nil {
				return []*gengo.TypeDecl{
					{
						Name: g.exportedGoName(id),
						Type: *identType,
					},
				}, nil
			}
		}
	}
	var fields gengo.Fields
	for name, prop := range s.Properties {
		goType, err := g.goType(name, prop)
		if err != nil {
			return nil, err
		}
		omitEmpty := true
		switch goType.Ident {
		case "float32", "float64":
			if strings.Contains(strings.ToLower(name), "price") {
				omitEmpty = false
			}
		}
		field := &gengo.Field{
			Name: g.exportedGoName(name),
			Type: *goType,
			Tag: gengo.Tag{
				{
					Encoding:  "json",
					Name:      name,
					OmitEmpty: omitEmpty,
				},
				{
					Encoding:  "yaml",
					Name:      name,
					OmitEmpty: omitEmpty,
				},
			},
		}
		fields = append(fields, field)
	}
	sort.Sort(fields)
	decls := []*gengo.TypeDecl{
		{
			Name: g.exportedGoName(id),
			Type: gengo.Type{
				Kind:   gengo.StructKind,
				Fields: fields,
			},
		},
	}

	for id, def := range s.Definitions {
		subDecls, err := g.goTypeDecls(id, def)
		if err != nil {
			return nil, err
		}
		decls = append(decls, subDecls...)
	}
	return decls, nil
}

const defPrefix = "#/definitions/"

func (g *generator) goType(name string, s *schema.Schema) (*gengo.Type, error) {
	switch s.Type {
	case "string", "integer", "number":
		return g.goIdentType(name, s.Type.(string))
	case "array":
		itemType, err := g.goType(name, s.Items)
		if err != nil {
			return nil, err
		}
		switch itemType.Kind {
		case gengo.IdentKind:
			return &gengo.Type{
				Kind:  gengo.ArrayKind,
				Ident: itemType.Ident,
			}, nil
		}
	case "object":
		return &gengo.Type{
			Kind:  gengo.IdentKind,
			Ident: "interface{}",
		}, nil
	}
	if s.Ref != nil && strings.HasPrefix(string(*s.Ref), defPrefix) {
		return g.goIdentType(name, strings.TrimPrefix(string(*s.Ref), defPrefix))
	}
	return nil, fmt.Errorf("fail to find Go type for %v", s)
}

func (g *generator) goIdentType(name, typ string) (*gengo.Type, error) {
	ident := ""
	switch typ {
	case "string":
		ident = "string"
	case "integer", "positive_int":
		ident = "int"
	case "boolean_int":
		ident = "BoolInt"
	case "number":
		name = strings.ToLower(name)
		if strings.Contains(name, "price") ||
			strings.Contains(name, "floor") ||
			strings.Contains(name, "ratio") {
			ident = "decimal.D"
		} else {
			ident = "float64"
		}
	default:
		ident = g.exportedGoName(typ)
	}
	return &gengo.Type{
		Kind:  gengo.IdentKind,
		Ident: ident,
	}, nil
}
