# Request

`ParseRequest` turns the arguments Compose passes into a `Request`:

```console
compose --project-name <NAME> <verb> [--key=value ...] <service>
```

| Field | From |
| --- | --- |
| `ProjectName` | `--project-name`, in either the space separated or `=` form |
| `Verb` | The first positional: `up`, `down`, `stop` or `metadata` |
| `Service` | The second positional, empty for `metadata` |
| `Options` | Every other `--key=value` flag |
| `Emitter` | Set by `Provider.Run` before the handler is called |

Options must use the `--key=value` form, which is the form Compose emits. A flag
given with no value at all reads as the boolean `true`. See
[../adr/01-00-00-options-use-the-key-value-form.md](../adr/01-00-00-options-use-the-key-value-form.md).

`ProjectName` is what a provider tags its resources with, so a later `down` can
find them again.

## Options

`Options` is `map[string][]string`: Compose emits one flag per element of an
array option, so a name repeats as often as the Compose file listed it. See
[../adr/01-00-01-options-are-always-slices.md](../adr/01-00-01-options-are-always-slices.md).

The accessors collapse that back to a value and report the mismatch when they
cannot:

| Accessor | Absent | Given once | Repeated |
| --- | --- | --- | --- |
| `String` | error | the value, error if empty | error |
| `StringDefault` | fallback | as `String` | error |
| `Bool`, `Int` | error | parsed, error if malformed | error |
| `BoolDefault`, `IntDefault` | fallback | as `Bool`, `Int` | error |
| `List` | `nil` | one element | every value, in order |
| `Has` | `false` | `true` | `true` |

`Provider.Run` fills in the defaults declared by the `Param` list before calling
the handler, so a declared option is read with `String`, `Bool` or `Int` rather
than their `Default` variants. The `Default` variants are for providers running
with `AllowUnknownOptions`, whose options have no `Param` to carry a default.

A repeated scalar is an error rather than last write wins, because silently
picking one of two configured values is how a service ends up pointing at the
wrong database.
