# 2. Prefer making backward compatible changes

Date: 2021-07-20

## Status

Proposed

Referenced by [3. Extract components with heavy dependencies from the core](0003-extract-components-with-heavy-dependencies-from-the-core.md)

Referenced by [4. Use separate GitHub organization for new packages](0004-use-separate-github-organization-for-new-packages.md)

## Context

Architecturally speaking Viper became a giant over the years: it hides a lot of complexity behind a simple interface.
That simple interface, however, is what makes Viper extremely popular.

## Decision

In order to keep the library useful to people, we should prefer making backward compatible changes to Viper, even between major releases.
This is not a hard rule forbiding breaking changes though: when it makes sense, breaking changes are allowed,
but keeping things backward compatible is a priority.

## Consequences

Although major versions allow breaking changes, a major release is no reason to break things that already work for a lot of people,
even if it might not be the best possible solution.

Instead of breaking things, introducing new interfaces should be the default way of fixing architectural problems,
leaving old interfaces intact.
