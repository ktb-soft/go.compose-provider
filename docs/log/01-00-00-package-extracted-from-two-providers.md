# 2026-09-08 — Package extracted from two existing providers

The protocol and invocation plumbing existed twice already, copied and diverged:

- `ktb-soft/secrets-injector` — `internal/protocol`, `internal/provider`
- `bkenks/collection_compose-extensions/src/infisical-secrets` — the same two
  packages, a generation behind

They disagreed on three things, each resolved by an ADR here: whether options
may be space separated
([adr/01-00-00](../adr/01-00-00-options-use-the-key-value-form.md)), whether an
option is a string or a slice
([adr/01-00-01](../adr/01-00-01-options-are-always-slices.md)), and whether
undeclared options are an error
([adr/01-00-02](../adr/01-00-02-unknown-options-are-rejected.md)).

Neither implementation covered the full protocol. `setenv` was missing from
both — each only ever emitted `rawsetenv` — and `debug` was missing from the
older one. Both are in the package.

Neither provider was migrated onto the package. That is separate work.
