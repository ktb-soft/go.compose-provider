package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// Handler runs one lifecycle verb. It reports progress through req.Emitter and
// returns an error to fail the service, which Compose renders as the reason.
type Handler func(ctx context.Context, req *Request) error

// Provider describes a Compose provider binary: the handlers Compose can call
// and the options each of them accepts.
type Provider struct {
	// Name is the binary name, reported by --version.
	Name string
	// Version is the release tag reported by --version, usually injected at
	// build time with -ldflags="-X main.version=...". It defaults to "dev".
	Version string
	// Description is the human readable summary reported by metadata.
	Description string

	// Up allocates the resource and hands its address to dependent services
	// with [Emitter.SetEnv]. It must be idempotent: a second run against an
	// existing resource has to set the same variables. Up is required.
	Up Handler
	// UpParams declares the options up accepts.
	UpParams []Param

	// Down releases every resource allocated for the project and service. A
	// nil Down makes down succeed without doing anything.
	Down Handler
	// DownParams declares the options down accepts.
	DownParams []Param

	// Stop pauses the resource without releasing it, so a later up resumes it.
	// Compose calls stop only when metadata declares it, so a nil Stop opts
	// out of the docker compose stop hook entirely.
	Stop Handler
	// StopParams declares the options stop accepts.
	StopParams []Param

	// AllowUnknownOptions passes options that no Param declares through to the
	// handler instead of rejecting them.
	AllowUnknownOptions bool
}

// Main runs p against os.Args and returns the process exit code, so a provider
// binary is:
//
//	func main() { os.Exit(compose.Main(provider)) }
//
// Failures are reported to Compose as an error message and exit code 1. The
// handler context is cancelled on SIGINT and SIGTERM, giving a provider the
// chance to release what it allocated when the user interrupts Compose.
func Main(p Provider) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := p.Run(ctx, os.Args[1:], os.Stdout); err != nil {
		NewEmitter(os.Stdout).Error("%v", err)
		return 1
	}
	return 0
}

// Run executes one invocation, writing the protocol stream to w. Args is
// os.Args[1:] when Compose runs the binary.
func (p Provider) Run(ctx context.Context, args []string, w io.Writer) error {
	if p.Up == nil {
		return errors.New("provider has no Up handler")
	}
	if len(args) == 1 && args[0] == "--version" {
		_, err := fmt.Fprintln(w, p.versionLine())
		return err
	}

	req, err := ParseRequest(args)
	if err != nil {
		return err
	}
	if req.Verb == VerbMetadata {
		return p.WriteMetadata(w)
	}

	handler, params := p.dispatch(req.Verb)
	applyDefaults(params, req.Options)
	if err := validateOptions(params, req.Options, p.AllowUnknownOptions); err != nil {
		return fmt.Errorf("service %q: %w", req.Service, err)
	}
	if handler == nil {
		return nil
	}

	req.Emitter = NewEmitter(w)
	if err := handler(ctx, req); err != nil {
		return err
	}
	return req.Emitter.Err()
}

func (p Provider) dispatch(verb Verb) (Handler, []Param) {
	switch verb {
	case VerbUp:
		return p.Up, p.UpParams
	case VerbDown:
		return p.Down, p.DownParams
	case VerbStop:
		return p.Stop, p.StopParams
	default:
		return nil, nil
	}
}

func (p Provider) versionLine() string {
	version := p.Version
	if version == "" {
		version = "dev"
	}
	if p.Name == "" {
		return version
	}
	return p.Name + " " + version
}
