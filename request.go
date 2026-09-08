package compose

import (
	"errors"
	"fmt"
	"strings"
)

// Verb is a lifecycle subcommand Compose invokes on a provider binary.
type Verb string

// The subcommands Compose invokes. A provider must accept up and down, may
// accept stop, and may accept metadata.
const (
	VerbUp       Verb = "up"
	VerbDown     Verb = "down"
	VerbStop     Verb = "stop"
	VerbMetadata Verb = "metadata"
)

const (
	composeCommand  = "compose"
	flagProjectName = "project-name"
)

// Request is one parsed provider invocation.
type Request struct {
	// ProjectName is the Compose project the service belongs to. Tag allocated
	// resources with it so down can find them again.
	ProjectName string
	// Verb is the lifecycle subcommand being run.
	Verb Verb
	// Service is the name of the provider service in the Compose file. It is
	// empty for VerbMetadata.
	Service string
	// Options is the provider.options block, one entry per flag name.
	Options Options
	// Emitter writes messages back to Compose. [Provider.Run] sets it before
	// calling a handler; [ParseRequest] leaves it nil.
	Emitter *Emitter
}

// ParseRequest parses the arguments Compose passes to a provider binary, which
// is os.Args[1:] when the binary runs. They take the form:
//
//	compose --project-name <NAME> <verb> [--key=value ...] <service>
//
// Options must use the --key=value form, which is the form Compose emits. A
// flag given without a value is read as the boolean true.
func ParseRequest(args []string) (*Request, error) {
	if len(args) == 0 || args[0] != composeCommand {
		return nil, fmt.Errorf("expected %q as the first argument", composeCommand)
	}

	parsed, err := splitFlags(args[1:])
	if err != nil {
		return nil, err
	}
	verb, service, err := readAction(parsed.positionals)
	if err != nil {
		return nil, err
	}

	return &Request{
		ProjectName: parsed.projectName,
		Verb:        verb,
		Service:     service,
		Options:     parsed.options,
	}, nil
}

type flagSet struct {
	projectName string
	options     Options
	positionals []string
}

// splitFlags separates --key=value flags from the positional verb and service.
// project-name is pulled out because Compose passes it space separated and it
// is not a provider option.
func splitFlags(args []string) (flagSet, error) {
	parsed := flagSet{options: Options{}}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			parsed.positionals = append(parsed.positionals, arg)
			continue
		}

		name, value, hasValue := strings.Cut(arg[2:], "=")
		if name == "" {
			return flagSet{}, fmt.Errorf("malformed flag %q", arg)
		}
		if name == flagProjectName {
			if !hasValue {
				if i+1 >= len(args) {
					return flagSet{}, fmt.Errorf("flag --%s requires a value", name)
				}
				i++
				value = args[i]
			}
			parsed.projectName = value
			continue
		}
		if !hasValue {
			value = "true"
		}
		parsed.options[name] = append(parsed.options[name], value)
	}
	return parsed, nil
}

// readAction reads the verb and service name out of the positional arguments.
func readAction(positionals []string) (Verb, string, error) {
	if len(positionals) == 0 {
		return "", "", errors.New("missing subcommand, expected one of up, down, stop, metadata")
	}
	if len(positionals) > 2 {
		return "", "", fmt.Errorf("unexpected arguments %v, options must use the --key=value form",
			positionals[2:])
	}

	verb := Verb(positionals[0])
	switch verb {
	case VerbUp, VerbDown, VerbStop, VerbMetadata:
	default:
		return "", "", fmt.Errorf("unsupported subcommand %q", verb)
	}

	var service string
	if len(positionals) == 2 {
		service = positionals[1]
	}
	if verb != VerbMetadata && service == "" {
		return "", "", fmt.Errorf("subcommand %q requires a service name", verb)
	}
	return verb, service, nil
}
