package compose

import (
	"encoding/json"
	"strings"
	"testing"
)

func decodeMetadata(t *testing.T, p Provider) map[string]any {
	t.Helper()

	var out strings.Builder
	if err := p.WriteMetadata(&out); err != nil {
		t.Fatalf("WriteMetadata() error = %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil {
		t.Fatalf("WriteMetadata() wrote invalid JSON %q: %v", out.String(), err)
	}
	return doc
}

func TestWriteMetadataDescribesEveryParam(t *testing.T) {
	doc := decodeMetadata(t, Provider{
		Description: "Manage services on AwesomeCloud",
		Up:          noopHandler,
		UpParams: []Param{
			{Name: "type", Description: "Database type", Required: true,
				Enum: []string{"mysql", "postgres"}},
			{Name: "size", Description: "Size in GB", Type: ParamInteger, Default: "10"},
		},
	})

	if doc["description"] != "Manage services on AwesomeCloud" {
		t.Errorf("description = %v", doc["description"])
	}

	params := doc["up"].(map[string]any)["parameters"].([]any)
	if len(params) != 2 {
		t.Fatalf("got %d up parameters, want 2", len(params))
	}

	first := params[0].(map[string]any)
	if first["type"] != string(ParamString) {
		t.Errorf("type of an untyped param = %v, want string", first["type"])
	}
	if first["required"] != true {
		t.Errorf("required = %v, want true", first["required"])
	}
	if _, ok := first["default"]; ok {
		t.Error("default is reported for a param that has none")
	}
	if enum := first["enum"].([]any); len(enum) != 2 {
		t.Errorf("enum = %v, want two values", enum)
	}

	second := params[1].(map[string]any)
	if second["type"] != string(ParamInteger) || second["default"] != "10" {
		t.Errorf("second param = %v", second)
	}
}

func TestWriteMetadataDeclaresStopOnlyWhenItIsImplemented(t *testing.T) {
	without := decodeMetadata(t, Provider{Up: noopHandler})
	if _, ok := without["stop"]; ok {
		t.Error("stop is declared by a provider that does not implement it")
	}
	if params := without["down"].(map[string]any)["parameters"]; params == nil {
		t.Error("down parameters = null, want an empty list")
	}

	with := decodeMetadata(t, Provider{Up: noopHandler, Stop: noopHandler})
	if _, ok := with["stop"]; !ok {
		t.Error("stop is missing from a provider that implements it")
	}
}

func TestValidateOptionsAcceptsADeclaredInvocation(t *testing.T) {
	params := []Param{
		{Name: "type", Required: true, Enum: []string{"mysql", "postgres"}},
		{Name: "size", Type: ParamInteger},
		{Name: "secret", Repeated: true},
	}
	options := Options{"type": {"mysql"}, "size": {"256"}, "secret": {"A", "B"}}

	if err := validateOptions(params, options, false); err != nil {
		t.Fatalf("validateOptions() error = %v", err)
	}
}

func TestValidateOptionsRejectsBadInvocations(t *testing.T) {
	params := []Param{
		{Name: "type", Required: true, Enum: []string{"mysql", "postgres"}},
		{Name: "size", Type: ParamInteger},
		{Name: "recursive", Type: ParamBoolean},
	}

	tests := map[string]Options{
		"missing required":  {"size": {"256"}},
		"unknown option":    {"type": {"mysql"}, "colour": {"blue"}},
		"value not in enum": {"type": {"sqlite"}},
		"not an integer":    {"type": {"mysql"}, "size": {"big"}},
		"not a boolean":     {"type": {"mysql"}, "recursive": {"maybe"}},
		"repeated scalar":   {"type": {"mysql", "postgres"}},
	}

	for name, options := range tests {
		t.Run(name, func(t *testing.T) {
			if err := validateOptions(params, options, false); err == nil {
				t.Fatalf("validateOptions(%v) = nil error, want a failure", options)
			}
		})
	}
}

func TestValidateOptionsReportsEveryProblemAtOnce(t *testing.T) {
	params := []Param{{Name: "type", Required: true}, {Name: "size", Type: ParamInteger}}
	options := Options{"size": {"big"}, "colour": {"blue"}}

	err := validateOptions(params, options, false)
	if err == nil {
		t.Fatal("validateOptions() = nil error, want a failure")
	}
	for _, want := range []string{"unknown option colour", "option type is required", "option size expects an integer"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestValidateOptionsPassesUnknownOptionsThroughWhenAllowed(t *testing.T) {
	options := Options{"colour": {"blue"}}

	if err := validateOptions(nil, options, true); err != nil {
		t.Fatalf("validateOptions() error = %v", err)
	}
	if err := validateOptions(nil, options, false); err == nil {
		t.Fatal("validateOptions() = nil error, want a failure")
	}
}
