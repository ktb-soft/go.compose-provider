package compose

import (
	"reflect"
	"testing"
)

func TestOptionsStringRequiresExactlyOneNonEmptyValue(t *testing.T) {
	options := Options{"type": {"mysql"}, "empty": {""}, "twice": {"a", "b"}}

	got, err := options.String("type")
	if err != nil || got != "mysql" {
		t.Errorf(`String("type") = %q, %v, want "mysql", nil`, got, err)
	}
	for _, name := range []string{"missing", "empty", "twice"} {
		if _, err := options.String(name); err == nil {
			t.Errorf("String(%q) = nil error, want a failure", name)
		}
	}
}

func TestOptionsStringDefaultFallsBackOnlyWhenAbsent(t *testing.T) {
	options := Options{"path": {"/secrets"}, "empty": {""}}

	got, err := options.StringDefault("missing", "/")
	if err != nil || got != "/" {
		t.Errorf(`StringDefault("missing", "/") = %q, %v, want "/", nil`, got, err)
	}
	got, err = options.StringDefault("path", "/")
	if err != nil || got != "/secrets" {
		t.Errorf(`StringDefault("path", "/") = %q, %v, want "/secrets", nil`, got, err)
	}
	if _, err := options.StringDefault("empty", "/"); err == nil {
		t.Error(`StringDefault("empty", "/") = nil error, want a failure`)
	}
}

func TestOptionsBool(t *testing.T) {
	options := Options{"recursive": {"true"}, "off": {"0"}, "junk": {"yes please"}}

	if got, err := options.Bool("recursive"); err != nil || !got {
		t.Errorf(`Bool("recursive") = %v, %v, want true, nil`, got, err)
	}
	if got, err := options.Bool("off"); err != nil || got {
		t.Errorf(`Bool("off") = %v, %v, want false, nil`, got, err)
	}
	if _, err := options.Bool("junk"); err == nil {
		t.Error(`Bool("junk") = nil error, want a failure`)
	}
	if got, err := options.BoolDefault("missing", true); err != nil || !got {
		t.Errorf(`BoolDefault("missing", true) = %v, %v, want true, nil`, got, err)
	}
}

func TestOptionsInt(t *testing.T) {
	options := Options{"size": {"256"}, "junk": {"big"}}

	if got, err := options.Int("size"); err != nil || got != 256 {
		t.Errorf(`Int("size") = %d, %v, want 256, nil`, got, err)
	}
	if _, err := options.Int("junk"); err == nil {
		t.Error(`Int("junk") = nil error, want a failure`)
	}
	if got, err := options.IntDefault("missing", 10); err != nil || got != 10 {
		t.Errorf(`IntDefault("missing", 10) = %d, %v, want 10, nil`, got, err)
	}
}

func TestOptionsInspection(t *testing.T) {
	options := Options{"secret": {"A", "B"}, "type": {"mysql"}}

	if !options.Has("secret") || options.Has("missing") {
		t.Error("Has() disagrees with the option map")
	}
	if got := options.Names(); !reflect.DeepEqual(got, []string{"secret", "type"}) {
		t.Errorf("Names() = %v, want [secret type]", got)
	}
	if got := options.List("secret"); !reflect.DeepEqual(got, []string{"A", "B"}) {
		t.Errorf(`List("secret") = %v, want [A B]`, got)
	}
	if got := options.List("missing"); got != nil {
		t.Errorf(`List("missing") = %v, want nil`, got)
	}
}
