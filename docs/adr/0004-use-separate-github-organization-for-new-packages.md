# 4. Use separate GitHub organization for new packages

Date: 2021-07-20

## Status

Proposed

References [2. Prefer making backward compatible changes](0002-prefer-making-backward-compatible-changes.md)

References [3. Extract components with heavy dependencies from the core](0003-extract-components-with-heavy-dependencies-from-the-core.md)

## Context

The core Viper package is under a personal GitHub account which makes collaborative development a bit difficult.

## Decision

Create new Go modules in the [go-viper](https://github.com/go-viper) organization.
Keep the core library under [Steve's personal account](https://github.com/spf13/viper) for backward compatibility purposes.

## Consequences

It'll be easier to create new modules and to add new functionality to Viper without having to add new dependencies to the core library.
