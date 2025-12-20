package foodora

import (
	"encoding/json"
	"testing"
)

func TestFlexibleInt(t *testing.T) {
	t.Parallel()

	var v struct {
		N FlexibleInt `json:"n"`
	}

	if err := json.Unmarshal([]byte(`{"n":123}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.N != 123 {
		t.Fatalf("got %d", v.N)
	}

	if err := json.Unmarshal([]byte(`{"n":"456"}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.N != 456 {
		t.Fatalf("got %d", v.N)
	}

	if err := json.Unmarshal([]byte(`{"n":12.9}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.N != 12 {
		t.Fatalf("got %d", v.N)
	}

	if err := json.Unmarshal([]byte(`{"n":null}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.N != 0 {
		t.Fatalf("got %d", v.N)
	}
}

func TestFlexibleString(t *testing.T) {
	t.Parallel()

	var v struct {
		S FlexibleString `json:"s"`
	}

	if err := json.Unmarshal([]byte(`{"s":"x"}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.S != "x" {
		t.Fatalf("got %q", v.S)
	}

	if err := json.Unmarshal([]byte(`{"s":123}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.S != "123" {
		t.Fatalf("got %q", v.S)
	}

	if err := json.Unmarshal([]byte(`{"s":true}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.S != "true" {
		t.Fatalf("got %q", v.S)
	}

	if err := json.Unmarshal([]byte(`{"s":null}`), &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.S != "" {
		t.Fatalf("got %q", v.S)
	}
}
