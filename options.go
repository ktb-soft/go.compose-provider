package compose

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
)

// Options is the provider.options block of a Compose service.
//
// Every name maps to a slice because Compose emits one flag per element of an
// array option, so a name repeats as often as the Compose file listed it.
type Options map[string][]string

// Has reports whether name was given at all.
func (o Options) Has(name string) bool { return len(o[name]) > 0 }

// Names returns every option name that was given, sorted.
func (o Options) Names() []string { return slices.Sorted(maps.Keys(o)) }

// List returns every value given for name, in command line order, and nil when
// the option is absent.
func (o Options) List(name string) []string { return o[name] }

// String returns the single value of a required option. It reports an error
// when the option is missing, empty, or given more than once.
func (o Options) String(name string) (string, error) {
	values := o[name]
	switch len(values) {
	case 0:
		return "", fmt.Errorf("option %s is required", name)
	case 1:
		if values[0] == "" {
			return "", fmt.Errorf("option %s is empty", name)
		}
		return values[0], nil
	default:
		return "", fmt.Errorf("option %s given %d times, want one", name, len(values))
	}
}

// StringDefault returns fallback when name is absent, and otherwise behaves
// like [Options.String], so an explicitly empty value is still an error.
func (o Options) StringDefault(name, fallback string) (string, error) {
	if !o.Has(name) {
		return fallback, nil
	}
	return o.String(name)
}

// Bool returns the single value of a required option parsed as a boolean,
// accepting the spellings strconv.ParseBool does.
func (o Options) Bool(name string) (bool, error) {
	value, err := o.String(name)
	if err != nil {
		return false, err
	}
	return parseBool(name, value)
}

// BoolDefault returns fallback when name is absent, and otherwise behaves like
// [Options.Bool].
func (o Options) BoolDefault(name string, fallback bool) (bool, error) {
	if !o.Has(name) {
		return fallback, nil
	}
	return o.Bool(name)
}

// Int returns the single value of a required option parsed as an integer.
func (o Options) Int(name string) (int, error) {
	value, err := o.String(name)
	if err != nil {
		return 0, err
	}
	return parseInt(name, value)
}

// IntDefault returns fallback when name is absent, and otherwise behaves like
// [Options.Int].
func (o Options) IntDefault(name string, fallback int) (int, error) {
	if !o.Has(name) {
		return fallback, nil
	}
	return o.Int(name)
}

func parseBool(name, value string) (bool, error) {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("option %s expects a boolean, got %q", name, value)
	}
	return parsed, nil
}

func parseInt(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("option %s expects an integer, got %q", name, value)
	}
	return parsed, nil
}
