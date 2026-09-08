package compose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

// Message types Compose reads from a provider's stdout stream.
const (
	typeInfo      = "info"
	typeError     = "error"
	typeDebug     = "debug"
	typeSetEnv    = "setenv"
	typeRawSetEnv = "rawsetenv"
)

// envNamePattern matches the variable names a container runtime accepts.
var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Emitter writes the newline delimited JSON messages Compose reads from a
// provider's stdout.
//
// Writes are fire and forget: the first failure is retained and every later
// message except [Emitter.Error] is dropped. [Provider.Run] returns that
// failure once the handler is done, so handlers need not check each call.
type Emitter struct {
	w   io.Writer
	err error
}

// NewEmitter returns an Emitter writing to w, which is os.Stdout when Compose
// runs the provider.
func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{w: w}
}

// Info reports a status update, which Compose renders as the service state in
// its progress UI.
func (e *Emitter) Info(format string, args ...any) {
	e.emit(typeInfo, fmt.Sprintf(format, args...))
}

// Debug reports detail Compose renders only under docker compose --verbose.
func (e *Emitter) Debug(format string, args ...any) {
	e.emit(typeDebug, fmt.Sprintf(format, args...))
}

// Error reports why the service failed. [Main] emits the error a handler
// returns, so handlers rarely call this directly.
//
// Unlike the other messages it is written even after an earlier write failed,
// so a failure is never silently swallowed.
func (e *Emitter) Error(format string, args ...any) {
	e.write(typeError, fmt.Sprintf(format, args...))
}

// SetEnv injects value into every service that declares depends_on for this
// one, under name prefixed with the provider service name. Name URL on service
// database arrives as DATABASE_URL.
func (e *Emitter) SetEnv(name, value string) {
	e.emitEnv(typeSetEnv, name, value)
}

// RawSetEnv injects value under name exactly as given, without the service
// name prefix [Emitter.SetEnv] applies.
//
// Compose overwrites any existing variable of the same name and logs a warning,
// so keeping raw names unique is the provider's responsibility.
func (e *Emitter) RawSetEnv(name, value string) {
	e.emitEnv(typeRawSetEnv, name, value)
}

// Err reports the first write failure, and nil when every message was written.
func (e *Emitter) Err() error { return e.err }

func (e *Emitter) emitEnv(kind, name, value string) {
	if e.err != nil {
		return
	}
	if !envNamePattern.MatchString(name) {
		e.err = fmt.Errorf("%q is not a valid environment variable name", name)
		return
	}
	e.emit(kind, name+"="+value)
}

func (e *Emitter) emit(kind, body string) {
	if e.err != nil {
		return
	}
	e.write(kind, body)
}

// write emits one protocol line. The body is JSON encoded, so a value holding
// quotes, backslashes or newlines cannot break the stream, and HTML escaping is
// off so a URL keeps its ampersands.
func (e *Emitter) write(kind, body string) {
	var line bytes.Buffer
	encoder := json.NewEncoder(&line)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(message{Type: kind, Message: body})
	if err == nil {
		_, err = e.w.Write(line.Bytes())
	}
	if err != nil && e.err == nil {
		e.err = err
	}
}
