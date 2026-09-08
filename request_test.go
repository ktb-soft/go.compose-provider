package compose

import (
	"reflect"
	"testing"
)

func TestParseRequestReadsAComposeUpInvocation(t *testing.T) {
	args := []string{
		"compose", "--project-name", "myproject", "up",
		"--type=mysql", "--size=256", "database",
	}

	req, err := ParseRequest(args)
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}

	if req.ProjectName != "myproject" {
		t.Errorf("ProjectName = %q, want myproject", req.ProjectName)
	}
	if req.Verb != VerbUp {
		t.Errorf("Verb = %q, want up", req.Verb)
	}
	if req.Service != "database" {
		t.Errorf("Service = %q, want database", req.Service)
	}
	want := Options{"type": {"mysql"}, "size": {"256"}}
	if !reflect.DeepEqual(req.Options, want) {
		t.Errorf("Options = %v, want %v", req.Options, want)
	}
	if req.Emitter != nil {
		t.Error("Emitter is set, want nil until the provider runs")
	}
}

func TestParseRequestCollectsRepeatedOptions(t *testing.T) {
	args := []string{
		"compose", "--project-name=myproject", "up",
		"--secret=A=one", "--secret=B=two", "database",
	}

	req, err := ParseRequest(args)
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}

	want := []string{"A=one", "B=two"}
	if got := req.Options.List("secret"); !reflect.DeepEqual(got, want) {
		t.Errorf("secret = %v, want %v", got, want)
	}
	if req.ProjectName != "myproject" {
		t.Errorf("ProjectName = %q, want myproject", req.ProjectName)
	}
}

func TestParseRequestReadsAValuelessFlagAsTrue(t *testing.T) {
	req, err := ParseRequest([]string{"compose", "up", "--recursive", "database"})
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}

	if got := req.Options.List("recursive"); !reflect.DeepEqual(got, []string{"true"}) {
		t.Errorf("recursive = %v, want [true]", got)
	}
	if req.Service != "database" {
		t.Errorf("Service = %q, want database", req.Service)
	}
}

func TestParseRequestAllowsMetadataWithoutAService(t *testing.T) {
	req, err := ParseRequest([]string{"compose", "metadata"})
	if err != nil {
		t.Fatalf("ParseRequest() error = %v", err)
	}

	if req.Verb != VerbMetadata {
		t.Errorf("Verb = %q, want metadata", req.Verb)
	}
	if req.Service != "" {
		t.Errorf("Service = %q, want empty", req.Service)
	}
}

func TestParseRequestRejectsBadInvocations(t *testing.T) {
	tests := map[string][]string{
		"no arguments":            {},
		"missing compose command": {"up", "database"},
		"missing subcommand":      {"compose", "--project-name", "myproject"},
		"unknown subcommand":      {"compose", "restart", "database"},
		"missing service":         {"compose", "up"},
		"space separated option":  {"compose", "up", "--type", "mysql", "database"},
		"malformed flag":          {"compose", "up", "--=mysql", "database"},
		"project name without value": {
			"compose", "up", "database", "--project-name",
		},
	}

	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseRequest(args); err == nil {
				t.Fatalf("ParseRequest(%v) = nil error, want a failure", args)
			}
		})
	}
}
