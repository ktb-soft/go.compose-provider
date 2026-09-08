package compose

import (
	"errors"
	"strings"
	"testing"
)

type failingWriter struct {
	writes int
}

func (w *failingWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, errors.New("stdout closed")
}

func TestEmitterWritesOneJSONLinePerMessage(t *testing.T) {
	var out strings.Builder
	emitter := NewEmitter(&out)

	emitter.Info("preparing %s", "mysql")
	emitter.Debug("resolving %d secrets", 3)
	emitter.SetEnv("URL", "https://awesomecloud.com/db:1234")
	emitter.RawSetEnv("SECRET_KEY", "xxx")
	emitter.Error("boom")

	want := []string{
		`{"type":"info","message":"preparing mysql"}`,
		`{"type":"debug","message":"resolving 3 secrets"}`,
		`{"type":"setenv","message":"URL=https://awesomecloud.com/db:1234"}`,
		`{"type":"rawsetenv","message":"SECRET_KEY=xxx"}`,
		`{"type":"error","message":"boom"}`,
	}
	got := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")

	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(got), len(want), out.String())
	}
	for i, line := range want {
		if got[i] != line {
			t.Errorf("line %d = %s, want %s", i, got[i], line)
		}
	}
	if err := emitter.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestEmitterKeepsHostileValuesOnOneLine(t *testing.T) {
	var out strings.Builder
	emitter := NewEmitter(&out)

	emitter.RawSetEnv("TOKEN", "a\"b\nc\\d\te<f>")

	if lines := strings.Count(out.String(), "\n"); lines != 1 {
		t.Fatalf("got %d newlines, want 1: %q", lines, out.String())
	}
	want := `{"type":"rawsetenv","message":"TOKEN=a\"b\nc\\d\te<f>"}` + "\n"
	if out.String() != want {
		t.Errorf("got %s, want %s", out.String(), want)
	}
}

func TestEmitterRejectsInvalidEnvNames(t *testing.T) {
	for _, name := range []string{"", "9LIVES", "with-dash", "with space", "A=B"} {
		t.Run(name, func(t *testing.T) {
			var out strings.Builder
			emitter := NewEmitter(&out)

			emitter.SetEnv(name, "value")

			if emitter.Err() == nil {
				t.Fatalf("Err() = nil, want an error for %q", name)
			}
			if out.String() != "" {
				t.Errorf("wrote %q, want nothing", out.String())
			}
		})
	}
}

func TestEmitterDropsMessagesAfterAFailedWrite(t *testing.T) {
	writer := &failingWriter{}
	emitter := NewEmitter(writer)

	emitter.Info("first")
	emitter.Info("second")

	if writer.writes != 1 {
		t.Errorf("attempted %d writes, want 1", writer.writes)
	}
	if emitter.Err() == nil {
		t.Fatal("Err() = nil, want the write failure")
	}
}

func TestEmitterReportsErrorsAfterAFailedWrite(t *testing.T) {
	writer := &failingWriter{}
	emitter := NewEmitter(writer)

	emitter.Info("first")
	emitter.Error("the reason the service failed")

	if writer.writes != 2 {
		t.Errorf("attempted %d writes, want 2", writer.writes)
	}
}
