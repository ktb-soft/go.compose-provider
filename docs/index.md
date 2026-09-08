# How go.compose-provider works

Compose runs a provider binary once per lifecycle event. Everything the binary
learns arrives on the command line, and everything it reports leaves on stdout
as newline delimited JSON. The package is three pieces around that shape.

```
os.Args ──▶ ParseRequest ──▶ validateOptions ──▶ Handler ──▶ Emitter ──▶ stdout
                  │                  │
             Request            []Param ──▶ WriteMetadata ──▶ stdout
```

## The lifecycle

Compose invokes:

```console
awesomecloud compose --project-name <NAME> up   --type=mysql "database"
awesomecloud compose --project-name <NAME> down             "database"
awesomecloud compose --project-name <NAME> stop             "database"
awesomecloud compose metadata
```

`up` allocates and must be idempotent: a second run against an existing resource
has to emit the same variables, because Compose relies on them to configure
dependent services. `down` releases everything held for that project and
service. `stop` pauses without releasing, so a later `up` resumes.

`stop` is opt-in. Compose calls it only when the `metadata` output declares a
`stop` block, so a provider with no `Stop` handler never sees it. See
[components/01-00-02-provider-runtime.md](components/01-00-02-provider-runtime.md).

## The pieces

| Piece | File | Does |
| --- | --- | --- |
| [Emitter](components/01-00-00-emitter.md) | `protocol.go` | Writes the JSON message stream Compose reads |
| [Request](components/01-00-01-request.md) | `request.go`, `options.go` | Parses the invocation and gives typed access to options |
| [Provider](components/01-00-02-provider-runtime.md) | `provider.go`, `metadata.go` | Dispatches verbs, applies and validates options, answers `metadata`, owns the exit code |

## The entry point

`Main` is the whole binary:

```go
func main() { os.Exit(compose.Main(provider)) }
```

It reads `os.Args`, cancels the handler context on `SIGINT` and `SIGTERM`, emits
any returned error as an `error` message, and exits 1.

`Provider.Run` is the same thing with the arguments and output stream passed in,
which is what tests use.
