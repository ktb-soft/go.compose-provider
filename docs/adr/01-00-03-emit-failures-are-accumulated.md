# Emit failures are accumulated, not returned

## Context

A handler emits many messages and each write can fail. Returning an error from
every `Info` call buries the provider's own logic in error checks that a caller
cannot usefully act on: if stdout is gone, Compose is not listening anyway.

## Decision

`Emitter` methods return nothing. The first failure is retained, later messages
are dropped, and `Provider.Run` returns `Emitter.Err()` once the handler is done,
so the failure still reaches the exit code.

`Error` is the exception. It attempts its write even after an earlier failure,
so the reason a service failed is never swallowed by the failure that caused it.

Invalid environment variable names are recorded the same way. `SetEnv("A-B", …)`
is a provider bug, and surfacing it through `Err` puts it on the same path as a
write failure without adding an error return to a call that has none.

## Consequences

Handler bodies read as the work they do. The one cost is that a bad variable name
surfaces after the handler returns rather than at the call, which the error
message names explicitly.

This rules out holding a single `json.Encoder` for the stream: an `Encoder`
becomes sticky after a failed `Encode` and refuses every later call, which would
suppress the `Error` message. Each message is encoded into a buffer and written
separately.
