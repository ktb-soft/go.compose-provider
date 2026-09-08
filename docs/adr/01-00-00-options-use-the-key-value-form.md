# Options use the --key=value form

## Context

Compose translates `provider.options` into command line flags and emits them as
`--key=value`. Two earlier providers disagreed on whether to also accept the
space separated `--key value` form for hand invocation.

The space separated form is ambiguous against the trailing service positional.
`--recursive database` is either the flag `recursive` with the value `database`
and no service, or a boolean `recursive` followed by the service `database`.
A parser that accepts both cannot tell them apart without knowing every option's
type, which is exactly what the parser runs before knowing.

## Decision

Options must use `--key=value`. A flag with no value at all is read as the
boolean `true`, which is unambiguous: `--recursive database` gives
`recursive=true` and the service `database`.

`--project-name` is the exception and accepts both forms, because Compose passes
it space separated.

## Consequences

The form Compose emits always parses. A human running the binary by hand gets an
error naming the `--key=value` form when they type `--path /secrets`, because
`/secrets` lands as an unexpected positional.
