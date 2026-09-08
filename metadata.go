package compose

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
)

// ParamType is the declared type of a provider option.
type ParamType string

// The parameter types Compose recognises. An empty [Param.Type] is read as
// ParamString.
const (
	ParamString  ParamType = "string"
	ParamInteger ParamType = "integer"
	ParamBoolean ParamType = "boolean"
)

// Param describes one option a provider accepts. Params are reported by the
// metadata subcommand and enforced before the matching handler runs.
type Param struct {
	// Name is the option name, without the leading dashes.
	Name string
	// Description is the human readable summary reported by metadata.
	Description string
	// Required rejects the invocation when the option is absent.
	Required bool
	// Type is the value type. An empty Type means ParamString.
	Type ParamType
	// Default is the value used when the option is absent. It is reported by
	// metadata and filled into [Request.Options] before the handler runs, so
	// the handler reads it through the plain accessor.
	Default string
	// Enum limits the accepted values. An empty Enum accepts anything.
	Enum []string
	// Repeated allows the option to be given more than once, which is how
	// Compose passes an array. It is enforced locally and not reported by
	// metadata, which has no field for it.
	Repeated bool
}

type paramJSON struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"`
	Default     string   `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type verbJSON struct {
	Parameters []paramJSON `json:"parameters"`
}

type metadataJSON struct {
	Description string    `json:"description"`
	Up          verbJSON  `json:"up"`
	Down        verbJSON  `json:"down"`
	Stop        *verbJSON `json:"stop,omitempty"`
}

// WriteMetadata writes the option schema Compose reads from the metadata
// subcommand to w, as indented JSON.
//
// The stop block is written only when [Provider.Stop] is set, because that
// block is what opts the provider into the docker compose stop hook.
func (p Provider) WriteMetadata(w io.Writer) error {
	doc := metadataJSON{
		Description: p.Description,
		Up:          describeParams(p.UpParams),
		Down:        describeParams(p.DownParams),
	}
	if p.Stop != nil {
		stop := describeParams(p.StopParams)
		doc.Stop = &stop
	}

	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding metadata: %w", err)
	}
	_, err = fmt.Fprintln(w, string(encoded))
	return err
}

func describeParams(params []Param) verbJSON {
	described := make([]paramJSON, 0, len(params))
	for _, param := range params {
		described = append(described, paramJSON{
			Name:        param.Name,
			Description: param.Description,
			Required:    param.Required,
			Type:        string(param.kind()),
			Default:     param.Default,
			Enum:        param.Enum,
		})
	}
	return verbJSON{Parameters: described}
}

// applyDefaults fills in the declared default for every option the invocation
// left out, so a default is stated once and the metadata output and the handler
// cannot drift apart.
func applyDefaults(params []Param, options Options) {
	for _, param := range params {
		if param.Default == "" || options.Has(param.Name) {
			continue
		}
		options[param.Name] = []string{param.Default}
	}
}

// validateOptions reports every problem with the given options at once, so a
// misconfigured service fails before the handler makes any remote call.
func validateOptions(params []Param, options Options, allowUnknown bool) error {
	var problems []string

	if !allowUnknown {
		problems = append(problems, unknownOptions(params, options)...)
	}
	for _, param := range params {
		if err := param.check(options); err != nil {
			problems = append(problems, err.Error())
		}
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("invalid options: %s", strings.Join(problems, "; "))
}

func unknownOptions(params []Param, options Options) []string {
	declared := make([]string, 0, len(params))
	for _, param := range params {
		declared = append(declared, param.Name)
	}
	slices.Sort(declared)

	var problems []string
	for _, name := range options.Names() {
		if _, found := slices.BinarySearch(declared, name); found {
			continue
		}
		if len(declared) == 0 {
			problems = append(problems, fmt.Sprintf("unknown option %s, none are accepted", name))
			continue
		}
		problems = append(problems, fmt.Sprintf("unknown option %s, expected one of: %s",
			name, strings.Join(declared, ", ")))
	}
	return problems
}

func (p Param) check(options Options) error {
	values := options.List(p.Name)
	if len(values) == 0 {
		if p.Required {
			return fmt.Errorf("option %s is required", p.Name)
		}
		return nil
	}
	if len(values) > 1 && !p.Repeated {
		return fmt.Errorf("option %s given %d times, want one", p.Name, len(values))
	}
	for _, value := range values {
		if err := p.checkValue(value); err != nil {
			return err
		}
	}
	return nil
}

func (p Param) checkValue(value string) error {
	var err error
	switch p.kind() {
	case ParamInteger:
		_, err = parseInt(p.Name, value)
	case ParamBoolean:
		_, err = parseBool(p.Name, value)
	}
	if err != nil {
		return err
	}
	if len(p.Enum) > 0 && !slices.Contains(p.Enum, value) {
		return fmt.Errorf("option %s is %q, expected one of: %s",
			p.Name, value, strings.Join(p.Enum, ", "))
	}
	return nil
}

func (p Param) kind() ParamType {
	if p.Type == "" {
		return ParamString
	}
	return p.Type
}
