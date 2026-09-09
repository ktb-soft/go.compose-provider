# go.compose-provider

Go package for speaking the Docker Compose provider language.

Docker Compose can hand a service off to an external binary instead of running
a container for it:

```yaml
services:
  app:
    image: myapp
    depends_on:
      - database

  database:
    provider:
      type: awesomecloud
      options:
        type: mysql
        size: 256
```

Compose then runs `awesomecloud compose --project-name <NAME> up --type=mysql
--size=256 "database"` and reads newline delimited JSON from its stdout to learn
what happened and which environment variables to inject into `app`.

This package is the binary side of that contract. It owns argument parsing, verb
dispatch, option validation, the `metadata` subcommand, the JSON message stream
and the exit code, so a provider is only its own logic.

## Install

```console
go get github.com/ktb-soft/go.compose-provider
```

## Use

```go
package main

import (
	"context"
	"os"

	compose "github.com/ktb-soft/go.compose-provider"
)

func main() {
	os.Exit(compose.Main(compose.Provider{
		Name:        "awesomecloud",
		Version:     version,
		Description: "Manage services on AwesomeCloud",
		Up:          up,
		UpParams: []compose.Param{
			{Name: "type", Description: "Database type", Required: true,
				Enum: []string{"mysql", "postgres"}},
			{Name: "size", Description: "Database size in GB",
				Type: compose.ParamInteger, Default: "10"},
		},
		Down: down,
	}))
}

func up(ctx context.Context, req *compose.Request) error {
	kind, err := req.Options.String("type")
	if err != nil {
		return err
	}
	size, err := req.Options.Int("size")
	if err != nil {
		return err
	}

	req.Emitter.Info("provisioning a %dGB %s", size, kind)
	req.Emitter.SetEnv("URL", "https://awesomecloud.com/db:1234")
	return nil
}

func down(ctx context.Context, req *compose.Request) error {
	req.Emitter.Info("releasing %s", req.Service)
	return nil
}
```

Put the binary on `PATH` under the name used as `provider.type`, and Compose
picks it up.

What the package handles for you:

- `compose up`, `compose down`, `compose stop` and `compose metadata`, plus
  `--version`
- Options parsed out of the command line, filled in from the defaults the
  `Param` list declares, then validated for required values, type, enum
  membership and repetition before a handler runs
- `metadata` generated from the same `Param` list, with the `stop` block
  declared only when a `Stop` handler is set, since that block is what opts a
  provider into the `docker compose stop` hook
- Errors turned into an `error` message on the stream and exit code 1
- `SIGINT` and `SIGTERM` cancelling the handler context

## Docs

- [docs/index.md](docs/index.md) — how the package works
- [docs/components/](docs/components) — the emitter, request parsing, the runtime
- [docs/adr/](docs/adr) — why the design is what it is

## Contributing

`mise run check` runs `go vet`, `golangci-lint` and the test suite. Tools are
pinned in `mise.toml`, so `mise install` gets you the same versions CI uses.

Releases are git tags, which is all a Go module needs:

```console
mise run release v0.1.0          # check, tag, push
mise run release:delete v0.1.0   # untag locally and on the remote
```

Both wrap the generic `tag:create` and `tag:delete` tasks, which only know how
to move a git tag: this repo supplies the semver rule, the `check` run and the
module proxy caveat. `tag:create` refuses a dirty tree, a tag that already
exists locally or on the remote, and unpushed commits.

`release:delete` is a last resort. The Go module proxy caches a version
permanently once anything fetches it, so a deleted tag is not withdrawn, and
re-tagging the same version against different code fails checksum verification
for anyone who already has it. Publish the next patch version instead.

## Reference

The protocol is the Compose extensibility specification, published by Docker at
[docker/compose/docs/extension.md](https://github.com/docker/compose/blob/main/docs/extension.md).

## Licence

MIT. See [LICENSE](LICENSE).
