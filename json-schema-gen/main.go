package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"h12.me/gengo"
	"h12.me/json-schema"
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
				if t, ok := m[field.Type.Ident]; ok && t.Type.Kind != gengo.IdentKind {
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

func (g *generator) goTypeDecls(id string, s *schema.Schema) ([]*gengo.TypeDecl, error) {
	if len(s.Properties) == 0 {
		if typ, ok := s.Type.(string); ok {
			if identType, err := g.goIdentType(typ); err == nil {
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
		goType, err := g.goType(prop)
		if err != nil {
			return nil, err
		}
		field := &gengo.Field{
			Name: g.exportedGoName(name),
			Type: *goType,
			Tag: gengo.Tag{
				{
					Encoding:  "json",
					Name:      name,
					OmitEmpty: true,
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

func (g *generator) goType(s *schema.Schema) (*gengo.Type, error) {
	switch s.Type {
	case "string", "integer", "number":
		return g.goIdentType(s.Type.(string))
	case "array":
		itemType, err := g.goType(s.Items)
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
		return g.goIdentType(strings.TrimPrefix(string(*s.Ref), defPrefix))
	}
	return nil, fmt.Errorf("fail to find Go type for %v", s)
}

func (g *generator) goIdentType(typ string) (*gengo.Type, error) {
	ident := ""
	switch typ {
	case "string":
		ident = "string"
	case "integer", "positive_int":
		ident = "int"
	case "boolean_int":
		ident = "BoolInt"
	case "number":
		ident = "float32"
	default:
		ident = g.exportedGoName(typ)
	}
	return &gengo.Type{
		Kind:  gengo.IdentKind,
		Ident: ident,
	}, nil
}
