# Enum is marshalled as a JSON array

## Context

The Compose extensibility specification describes the `enum` property of a
metadata parameter as "list of possible values supported by the parameter
separated by `,`", which reads as a comma separated string, while every worked
example in the wild writes it as a JSON array.

## Decision

`Param.Enum` is `[]string` and marshals as a JSON array:

```json
{"name": "type", "type": "string", "enum": ["mysql", "postgres"]}
```

## Consequences

This matches the shape Compose parses and the shape existing providers emit. If
a Compose release turns out to want the comma separated string, only
`describeParams` changes — the `Param` API stays as it is.

Enum membership is enforced locally regardless of how it is reported, so a bad
value fails whether or not Compose reads the metadata.
