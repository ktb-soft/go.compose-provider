package compose_test

import (
	"context"
	"fmt"
	"os"

	compose "github.com/ktb-soft/go.compose-provider"
)

// A provider binary hands its Provider to Main, which owns argument parsing,
// verb dispatch, the metadata subcommand and the exit code.
func ExampleMain() {
	provider := compose.Provider{
		Name:        "awesomecloud",
		Version:     "1.2.3",
		Description: "Manage services on AwesomeCloud",
		Up:          up,
		UpParams: []compose.Param{
			{
				Name:        "type",
				Description: "Database type",
				Required:    true,
				Enum:        []string{"mysql", "postgres"},
			},
			{
				Name:        "size",
				Description: "Database size in GB",
				Type:        compose.ParamInteger,
				Default:     "10",
			},
		},
		Down: down,
	}

	os.Exit(compose.Main(provider))
}

func up(_ context.Context, req *compose.Request) error {
	kind, err := req.Options.String("type")
	if err != nil {
		return err
	}
	size, err := req.Options.Int("size")
	if err != nil {
		return err
	}

	req.Emitter.Info("provisioning a %dGB %s for %s", size, kind, req.ProjectName)
	req.Emitter.SetEnv("URL", "https://awesomecloud.com/db:1234")
	return nil
}

func down(_ context.Context, req *compose.Request) error {
	req.Emitter.Info("releasing %s", req.Service)
	return nil
}

// Run executes a single invocation, which is what Main does once it has read
// os.Args. It is the entry point to reach for in tests.
func ExampleProvider_Run() {
	provider := compose.Provider{
		Up: func(_ context.Context, req *compose.Request) error {
			req.Emitter.Info("provisioning %s", req.Service)
			req.Emitter.SetEnv("URL", "https://awesomecloud.com/db:1234")
			return nil
		},
		UpParams: []compose.Param{{Name: "type", Required: true}},
	}

	args := []string{"compose", "--project-name", "myproject", "up", "--type=mysql", "database"}
	if err := provider.Run(context.Background(), args, os.Stdout); err != nil {
		fmt.Println(err)
	}

	// Output:
	// {"type":"info","message":"provisioning database"}
	// {"type":"setenv","message":"URL=https://awesomecloud.com/db:1234"}
}
