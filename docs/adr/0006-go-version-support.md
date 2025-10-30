# 6. Go version support

Date: 2021-09-16

## Status

Proposed

## Context

From time to time new features are released in the Go language.
Relying on those features means dropping support for older Go versions.

## Decision

Follow the [Go release policy](https://golang.org/doc/devel/release#policy) and support the last two major versions of Go.

## Consequences

Support for older Go versions will happen every 6 months according to the Go release cycle.
