# Declared defaults are applied, not just advertised

## Context

A `Param` can carry a `Default`, which the `metadata` subcommand reports so
Compose and anyone reading the output know what an omitted option resolves to.

Reporting it is not the same as using it. If the handler also has to write the
fallback — `req.Options.IntDefault("size", 10)` next to `Default: "10"` — the
value is stated twice in two forms, and nothing keeps them in step. The metadata
output can advertise a number the binary does not use.

## Decision

`Provider.Run` fills in every declared default that the invocation left out
before validating, so `Default` is stated once and the handler reads it through
the plain `String`, `Bool` or `Int` accessor.

Defaults are applied before validation, so a default that violates its own
`Enum` or `Type` fails immediately rather than on the first invocation that
happens to omit the option.

The `StringDefault`, `BoolDefault` and `IntDefault` accessors stay, for
providers running with `AllowUnknownOptions` whose options have no `Param` to
carry a default.

## Consequences

Every `Param` field is now declared once and used twice: reported by `metadata`
and enforced, or applied, before the handler runs.

A `Param` that is both `Required` and defaulted never fails the required check.
That combination is contradictory and is left to the provider author.
