# Emitter

`Emitter` writes the messages Compose reads from a provider's stdout: one JSON
object per line, each with a `type` and a `message`.

| Method | `type` | Compose does |
| --- | --- | --- |
| `Info` | `info` | Renders the message as the service state in the progress UI |
| `Debug` | `debug` | Renders it only under `docker compose --verbose` |
| `Error` | `error` | Renders it as the reason the service failed |
| `SetEnv` | `setenv` | Injects the variable into dependent services, prefixed with the provider service name |
| `RawSetEnv` | `rawsetenv` | Injects the variable under exactly the given name |

## Prefixed and raw variables

`SetEnv("URL", …)` on a service named `database` reaches dependent services as
`DATABASE_URL`. The prefix is what keeps two providers in one project from
colliding.

`RawSetEnv` skips it, for the case where an application demands an exact name.
Compose overwrites any existing variable of that name and logs a warning, so
uniqueness is the provider's problem. Providers that are not linked by
`depends_on` can run concurrently, so two of them emitting the same raw name
produce a value that is not deterministic.

Both reject a name that is not a valid environment variable name — matching
`^[A-Za-z_][A-Za-z0-9_]*$` — rather than emitting a line the runtime would
mangle.

## Failures

Writes are fire and forget. The first failure is retained, every later message
except `Error` is dropped, and `Provider.Run` returns the failure once the
handler is done. `Error` is written even after an earlier failure, so the reason
a service failed is never swallowed by the failure that caused it. See
[../adr/01-00-03-emit-failures-are-accumulated.md](../adr/01-00-03-emit-failures-are-accumulated.md).

Message bodies are JSON encoded, so a secret holding a quote, a backslash or a
newline cannot break the stream. HTML escaping is off, so a URL keeps its
ampersands intact.
