# Options are always slices

## Context

Compose emits one flag per element of an array option, so a Compose file
declaring three secrets produces three `--secret=` flags. A `map[string]string`
loses two of them to last write wins, silently.

## Decision

`Options` is `map[string][]string`. The typed accessors — `String`, `Bool`,
`Int` and their `Default` variants — reject an option that appears more than
once, and `List` returns every value for options that are meant to repeat.

`Param.Repeated` declares which options may legitimately appear more than once,
and validation enforces it before the handler runs.

## Consequences

A repeated scalar fails loudly instead of resolving to whichever value came
last. Providers that take arrays get them intact.
