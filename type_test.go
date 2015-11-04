package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestMeta(t *testing.T) {
	f, err := os.Open("testdata/meta-schema.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var s Schema
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		t.Fatal(err)
	}
	out, err := os.Create("testdata/meta-schema.out.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(out).Encode(&s); err != nil {
		t.Fatal(err)
	}
}

func TestOpenRTB(t *testing.T) {
	f, err := os.Open("testdata/request.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var s Schema
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}
