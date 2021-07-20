# 5. Deprecate setters in favor of functional options during initialization

Date: 2021-07-20

## Status

Proposed

## Context

The Viper struct currently acts as a facade for reading, writing and watching configuration for changes.
Some of the configuration parameters can be changed runtime using setters which often lead to issues
with concurrent activities.

## Decision

Deprecate setters in favor of using functional options for configuring Viper when it's initialized.

Drop setters in Viper 2.

## Consequences

Since Viper's interface is usually invoked from a lot of places,
moving configuration to the place where it is initialized makes using Viper safer
(ie. someone can't just randomly call `Set` when they are only supposed to call `Get*`).

This change will also clarify what roles Viper can be used in and
makes the separation of internal components easier based on these roles.
