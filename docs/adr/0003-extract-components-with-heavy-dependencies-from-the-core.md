# 3. Extract components with heavy dependencies from the core

Date: 2021-07-20

## Status

Proposed

References [2. Prefer making backward compatible changes](0002-prefer-making-backward-compatible-changes.md)

Referenced by [4. Use separate GitHub organization for new packages](0004-use-separate-github-organization-for-new-packages.md)

## Context

Viper (v1) currently imports a bunch of external dependencies (for encoding/decoding, remote stores, etc)
that make the library itself quite a heavy dependency.

## Decision

Move components with external dependencies out of the core to separate packages.

## Consequences

Viper 1 will have to continue importing all of these packages to maintain backwards compatibility.

Viper 2 (and future versions) on the other hand can break backwards compatibility and require users to import the required packages.
