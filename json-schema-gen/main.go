package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"h12.me/gengo"
	"h12.me/json-schema"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("json-schema-gen [filenames...]")
		return
	}
	var decls []*gengo.TypeDecl
	for _, filename := range os.Args[1:] {
		ds, err := collectDecls(filename)
		if err != nil {
			log.Fatal(err)
		}
		decls = append(decls, ds...)
	}
	codeFile := gengo.File{
		PackageName: "openrtb",
		TypeDecls:   filterDecls(decls),
	}
	codeFile.Fprint(os.Stdout)
}

func collectDecls(filename string) ([]*gengo.TypeDecl, error) {
	var s schema.Schema
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	decls, err := goTypeDecls(s.ID, &s)
	if err != nil {
		return nil, err
	}
	return decls, nil
}

func filterDecls(decls []*gengo.TypeDecl) (res []*gengo.TypeDecl) {
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

func goTypeDecls(id string, s *schema.Schema) ([]*gengo.TypeDecl, error) {
	if len(s.Properties) == 0 {
		if typ, ok := s.Type.(string); ok {
			if identType, err := goIdentType(typ); err == nil {
				return []*gengo.TypeDecl{
					{
						Name: exportedGoName(id),
						Type: *identType,
					},
				}, nil
			}
		}
	}
	var fields []*gengo.Field
	for name, prop := range s.Properties {
		goType, err := goType(prop)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &gengo.Field{
			Name: exportedGoName(name),
			Type: *goType,
			Tag: &gengo.Tag{
				Parts: []*gengo.TagPart{
					{
						Encoding:  "json",
						Name:      name,
						OmitEmpty: true,
					},
				},
			},
		})
	}
	decls := []*gengo.TypeDecl{
		{
			Name: exportedGoName(id),
			Type: gengo.Type{
				Kind:   gengo.StructKind,
				Fields: fields,
			},
		},
	}

	for id, def := range s.Definitions {
		subDecls, err := goTypeDecls(id, def)
		if err != nil {
			return nil, err
		}
		decls = append(decls, subDecls...)
	}
	return decls, nil
}

const defPrefix = "#/definitions/"

func goType(s *schema.Schema) (*gengo.Type, error) {
	switch s.Type {
	case "string", "integer", "number":
		return goIdentType(s.Type.(string))
	case "array":
		itemType, err := goType(s.Items)
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
		return goIdentType(strings.TrimPrefix(string(*s.Ref), defPrefix))
	}
	return nil, fmt.Errorf("fail to find Go type for %v", s)
}

func exportedGoName(s string) string {
	s = snakeToCamel(s)
	if strings.HasSuffix(s, "Id") {
		return strings.TrimSuffix(s, "Id") + "ID"
	}
	return s
}
func snakeToCamel(s string) string {
	ss := strings.Split(s, "_")
	for i := range ss {
		ss[i] = strings.Title(ss[i])
	}
	return strings.Join(ss, "")
}

func goIdentType(typ string) (*gengo.Type, error) {
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
		ident = exportedGoName(typ)
	}
	return &gengo.Type{
		Kind:  gengo.IdentKind,
		Ident: ident,
	}, nil
}
