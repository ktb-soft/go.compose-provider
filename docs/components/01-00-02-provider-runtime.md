# Provider runtime

`Provider` is the description of a binary: the handlers Compose can call, and
the options each accepts.

```go
compose.Provider{
	Name:        "awesomecloud",
	Version:     version,
	Description: "Manage services on AwesomeCloud",
	Up:          up,
	UpParams:    []compose.Param{{Name: "type", Required: true}},
	Down:        down,
}
```

`Run` does, in order:

1. Rejects a `Provider` with no `Up` handler, which is a programming error
2. Answers `--version` with `<name> <version>`, defaulting the version to `dev`
3. Parses the invocation
4. Answers `metadata` from the `Param` lists and stops there
5. Fills in the defaults declared by the `Param` list for the verb, then
   validates the options against it
6. Attaches an `Emitter` and calls the handler
7. Returns the handler's error, or any failure the emitter accumulated

A nil `Down` or `Stop` succeeds without doing anything, which is right for a
provider that allocates nothing remotely.

## Params

One `Param` per option, declared once and used twice: it is what `metadata`
reports, and it is what the invocation is checked against — and completed from —
before a handler touches the network. A default stated in a `Param` is the only
place it is stated, so what `metadata` advertises is what the handler reads.

| Field | Effect |
| --- | --- |
| `Name` | The option name, without dashes |
| `Description` | Reported by `metadata` |
| `Required` | Absence is an error |
| `Type` | `ParamString` (the default), `ParamInteger` or `ParamBoolean`; the value is parsed to check it |
| `Default` | Reported by `metadata` and filled into `Options` when the option is absent |
| `Enum` | Limits the accepted values |
| `Repeated` | Allows the option more than once, which is how Compose passes an array |

`Repeated` is local only. The metadata format has no field for it.

Validation reports every problem in one error rather than the first, so a
misconfigured service is fixed in one pass.

Options that no `Param` declares are rejected. `AllowUnknownOptions` passes them
through to the handler instead, for a provider whose options are not known ahead
of time. See
[../adr/01-00-02-unknown-options-are-rejected.md](../adr/01-00-02-unknown-options-are-rejected.md).

## The stop hook

`WriteMetadata` emits a `stop` block only when `Stop` is set. That block is what
opts a provider into `docker compose stop`: Compose silently skips `stop` for
providers that do not advertise it, which is what keeps the hook backward
compatible with providers written before it existed.

The `--timeout` flag of `docker compose stop` does not apply to provider hooks,
so a `Stop` handler manages its own shutdown duration.
