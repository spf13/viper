# 7. Drop writing support

Date: 2021-09-22

## Status

Proposed

References [2. Prefer making backward compatible changes](0002-prefer-making-backward-compatible-changes.md)

## Context

The number one source of issues for Viper comes from the fact that it supports both reading and writing.
It causes concurrency issues and has lots of inconsistencies.

## Decision

Drop file writing support from Viper in v2.

## Consequences

This is going to be a major breaking change in the library, but it will make maintenance significantly easier.
