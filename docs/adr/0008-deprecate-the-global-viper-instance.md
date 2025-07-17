# 8. Deprecate the global Viper instance

Date: 2021-09-23

## Status

Proposed

References [5. Deprecate setters in favor of functional options during initialization](0005-deprecate-setters-in-favor-of-functional-options-during-initialization.md)

## Context

With the deprecation of setters in favor of functional options, it becomes almost impossible to get away with instantiating Viper.
In addition to that, people should be discouraged from accessing a global Viper instance.

## Decision

Deprecate the global Viper instance and the global access functions.

## Consequences

People will still be able to create a global instance of their own,
but instantiating a custom Viper instance will become the primary solution for using Viper.
