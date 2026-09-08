package compose

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func noopHandler(context.Context, *Request) error { return nil }

func TestRunCallsTheHandlerForItsVerb(t *testing.T) {
	var seen *Request
	provider := Provider{
		Up: func(_ context.Context, req *Request) error {
			seen = req
			req.Emitter.Info("provisioning")
			return nil
		},
		UpParams: []Param{{Name: "type", Required: true}},
	}

	var out strings.Builder
	args := []string{"compose", "--project-name", "myproject", "up", "--type=mysql", "database"}
	if err := provider.Run(t.Context(), args, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if seen == nil {
		t.Fatal("the up handler was not called")
	}
	if seen.Service != "database" || seen.ProjectName != "myproject" {
		t.Errorf("handler saw service %q in project %q", seen.Service, seen.ProjectName)
	}
	if seen.Emitter == nil {
		t.Error("handler saw a nil Emitter")
	}
	if want := `{"type":"info","message":"provisioning"}` + "\n"; out.String() != want {
		t.Errorf("wrote %q, want %q", out.String(), want)
	}
}

func TestRunSucceedsWithoutAHandlerForDownAndStop(t *testing.T) {
	provider := Provider{Up: noopHandler}

	for _, verb := range []string{"down", "stop"} {
		t.Run(verb, func(t *testing.T) {
			var out strings.Builder
			args := []string{"compose", "--project-name", "myproject", verb, "database"}
			if err := provider.Run(t.Context(), args, &out); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if out.String() != "" {
				t.Errorf("wrote %q, want nothing", out.String())
			}
		})
	}
}

func TestRunAnswersTheMetadataSubcommand(t *testing.T) {
	provider := Provider{Description: "Manage services on AwesomeCloud", Up: noopHandler}

	var out strings.Builder
	if err := provider.Run(t.Context(), []string{"compose", "metadata"}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil {
		t.Fatalf("metadata is not JSON: %v", err)
	}
	if doc["description"] != "Manage services on AwesomeCloud" {
		t.Errorf("description = %v", doc["description"])
	}
}

func TestRunValidatesOptionsBeforeCallingTheHandler(t *testing.T) {
	called := false
	provider := Provider{
		Up:       func(context.Context, *Request) error { called = true; return nil },
		UpParams: []Param{{Name: "type", Required: true}},
	}

	var out strings.Builder
	args := []string{"compose", "up", "--colour=blue", "database"}
	err := provider.Run(t.Context(), args, &out)

	if err == nil {
		t.Fatal("Run() = nil error, want a failure")
	}
	if called {
		t.Error("the up handler ran despite invalid options")
	}
	if !strings.Contains(err.Error(), `service "database"`) {
		t.Errorf("error %q does not name the service", err)
	}
}

func TestRunReturnsTheHandlerError(t *testing.T) {
	wantErr := errors.New("the cloud said no")
	provider := Provider{Up: func(context.Context, *Request) error { return wantErr }}

	var out strings.Builder
	err := provider.Run(t.Context(), []string{"compose", "up", "database"}, &out)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestRunReturnsAFailedEmit(t *testing.T) {
	provider := Provider{
		Up: func(_ context.Context, req *Request) error {
			req.Emitter.SetEnv("not a name", "value")
			return nil
		},
	}

	var out strings.Builder
	err := provider.Run(t.Context(), []string{"compose", "up", "database"}, &out)

	if err == nil {
		t.Fatal("Run() = nil error, want the emit failure")
	}
}

func TestRunReportsTheVersion(t *testing.T) {
	var out strings.Builder
	provider := Provider{Name: "awesomecloud", Version: "1.2.3", Up: noopHandler}
	if err := provider.Run(t.Context(), []string{"--version"}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out.String() != "awesomecloud 1.2.3\n" {
		t.Errorf("wrote %q, want %q", out.String(), "awesomecloud 1.2.3\n")
	}

	out.Reset()
	unversioned := Provider{Name: "awesomecloud", Up: noopHandler}
	if err := unversioned.Run(t.Context(), []string{"--version"}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if out.String() != "awesomecloud dev\n" {
		t.Errorf("wrote %q, want %q", out.String(), "awesomecloud dev\n")
	}
}

func TestRunRejectsAProviderWithoutAnUpHandler(t *testing.T) {
	var out strings.Builder
	err := Provider{}.Run(t.Context(), []string{"compose", "up", "database"}, &out)

	if err == nil {
		t.Fatal("Run() = nil error, want a failure")
	}
}

func TestRunFillsInDeclaredDefaults(t *testing.T) {
	var seen Options
	provider := Provider{
		Up: func(_ context.Context, req *Request) error {
			seen = req.Options
			return nil
		},
		UpParams: []Param{
			{Name: "size", Type: ParamInteger, Default: "10"},
			{Name: "type", Default: "mysql"},
		},
	}

	var out strings.Builder
	args := []string{"compose", "up", "--type=postgres", "database"}
	if err := provider.Run(t.Context(), args, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got, err := seen.Int("size"); err != nil || got != 10 {
		t.Errorf(`Int("size") = %d, %v, want 10, nil`, got, err)
	}
	if got, err := seen.String("type"); err != nil || got != "postgres" {
		t.Errorf(`String("type") = %q, %v, want "postgres", nil`, got, err)
	}
}

func TestRunValidatesDeclaredDefaults(t *testing.T) {
	provider := Provider{
		Up:       noopHandler,
		UpParams: []Param{{Name: "size", Type: ParamInteger, Default: "ten"}},
	}

	var out strings.Builder
	err := provider.Run(t.Context(), []string{"compose", "up", "database"}, &out)

	if err == nil {
		t.Fatal("Run() = nil error, want the invalid default to fail")
	}
}
