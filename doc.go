// Package compose implements the Docker Compose provider protocol, the
// contract a binary named by a service's provider.type attribute must satisfy.
//
// Compose runs the binary once per lifecycle event, passing the project name,
// the service name, and the provider.options block translated into flags:
//
//	awesomecloud compose --project-name myproject up --type=mysql "database"
//
// The binary answers on stdout with newline delimited JSON messages that report
// progress and hand environment variables back to the services that depend on
// it.
//
// A provider is described by a [Provider] value and run by [Main], which owns
// argument parsing, verb dispatch, option validation, the metadata subcommand
// and the exit code:
//
//	func main() {
//		os.Exit(compose.Main(compose.Provider{
//			Name:        "awesomecloud",
//			Description: "Manage services on AwesomeCloud",
//			Up:          runUp,
//			UpParams: []compose.Param{
//				{Name: "type", Description: "Database type", Required: true},
//			},
//		}))
//	}
//
//	func runUp(ctx context.Context, req *compose.Request) error {
//		kind, err := req.Options.String("type")
//		if err != nil {
//			return err
//		}
//		req.Emitter.Info("provisioning %s", kind)
//		req.Emitter.SetEnv("URL", "https://awesomecloud.com/db:1234")
//		return nil
//	}
package compose
