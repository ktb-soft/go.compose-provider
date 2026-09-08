# Unknown options are rejected by default

## Context

Compose passes whatever `provider.options` the Compose file declares. A typo
produces an option the provider does not know. Ignoring it means the provider
runs with a silently different configuration than the file describes.

One earlier provider deliberately shipped no metadata so Compose would pass
every option through unfiltered, because its option set depends on which backend
the file selects and is not known ahead of time.

## Decision

An option that no `Param` declares is an error. `Provider.AllowUnknownOptions`
turns that off and passes them through to the handler.

## Consequences

The common case catches typos at the point of failure, with an error naming the
options the provider does accept. The dynamic case is still expressible, at the
cost of one explicit opt-out rather than an accident.

A provider that declares no `Param` for a verb accepts no options for that verb.
That is the correct reading for `down` and `stop`, and it is a loud failure for
an author who forgot to declare `UpParams`.
