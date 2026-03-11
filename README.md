# GETL

Portuguese (Brazil) version: [docs/README.pt-BR.md](./docs/README.pt-BR.md)

## Table of Contents

- [Overview](#overview)
- [Current Product Scope](#current-product-scope)
- [Current Operational State](#current-operational-state)
- [Core Capabilities](#core-capabilities)
- [Architecture Overview](#architecture-overview)
- [Repository Layout](#repository-layout)
- [Installation Notes](#installation-notes)
- [Primary Commands](#primary-commands)
- [Configuration Model](#configuration-model)
- [PostgreSQL and Catalog Sync Use Case](#postgresql-and-catalog-sync-use-case)
- [Current Role in the Ecosystem](#current-role-in-the-ecosystem)
- [Current Limitations](#current-limitations)
- [Screenshots](#screenshots)

## Overview

`GETL` is the Kubex ETL and synchronization utility.

It is designed to move data between heterogeneous sources and destinations through configurable pipelines, while remaining practical enough to be used in real repository-to-runtime workflows.

At the current stage, `GETL` is not just a generic ETL idea. It is now materially used as part of the Sankhya catalog ingestion path that supports metadata-driven BI generation in `GNyx`.

## Current Product Scope

Current practical areas include:

- configurable ETL and sync execution
- SQL-oriented source and destination flows
- database and file-oriented ingestion paths
- full-refresh or batch-style materialization
- utility support for heterogeneous systems
- reusable Go code plus CLI-oriented execution paths

It also includes a broader long-term surface around incremental sync, multiple backends, and richer synchronization strategies.

## Current Operational State

Operationally relevant truths today:

- `GETL` is already being used to load Sankhya metadata CSVs into PostgreSQL
- the current real target schema for that front is `sankhya_catalog`
- config files can now expand environment variables, which made versioned sync configs practical
- the current repository is valuable both as a tool and as an importable dependency

## Core Capabilities

Current concrete capabilities include:

- configurable extraction and load flows
- SQL-backed ingestion and load helpers
- destination table materialization
- batch load execution from CSV-oriented inputs
- broader support for heterogeneous data movement patterns
- config loading with env expansion

## Architecture Overview

`GETL` is organized around:

- CLI entrypoints
- configuration and sync orchestration
- SQL/data utility layers
- extraction and transformation support
- shared ETL-oriented types

## Repository Layout

```text
cmd/                    CLI entrypoints
sql/                    SQL-oriented ETL paths
sync/                   synchronization logic
utils/                  loading and supporting helpers
extr/                   extraction-related code
etypes/                 ETL-oriented shared types
```

## Installation Notes

Important build constraint:

- because `GETL` depends on `godror` for Oracle support, some build environments require `CGO=1`

Example:

```bash
CGO=1 go build ./...
```

This matters not only when building `GETL` directly, but also when another project imports `GETL` transitively.

## Primary Commands

Build:

```bash
go build ./...
```

Run tests:

```bash
go test ./...
```

The exact operational command surface can vary by flow, but `GETL` is already being consumed through higher-level orchestration from `GNyx`.

## Configuration Model

`GETL` is configuration-driven.

Recent practical improvement:

- environment variable expansion in config loading now works, which means repository-committed sync manifests no longer need to hardcode local DSNs or machine-specific values

That made it viable to keep reusable sync configs in version control while still running them across different environments.

## PostgreSQL and Catalog Sync Use Case

A key real use case now exists:

- Sankhya BI metadata CSVs are loaded into PostgreSQL
- target schema: `sankhya_catalog`
- registry/governance stays in `Domus`
- orchestration command currently lives in `GNyx`

In practical terms:

- `GETL` handles ingestion/materialization
- `Domus` hosts the active PostgreSQL runtime and external metadata registry
- `GNyx` orchestrates the domain-specific sync flow

This is a strong example of `GETL` being used as an actual ecosystem tool rather than a standalone experiment.

## Current Role in the Ecosystem

Today `GETL` is especially relevant for:

- metadata ingestion fronts
- catalog loading into PostgreSQL
- future broader ETL and sync scenarios across Kubex projects

Its role increased materially once the metadata-driven BI proof of concept became a real `GNyx` feature path.

## Current Limitations

Current limitations include:

- ecosystem use is still concentrated in a few concrete flows rather than every theoretical capability
- the overall surface is broader than the currently battle-tested slice
- `godror` and CGO requirements add practical build constraints in some environments

## Screenshots

Placeholder suggestions:

- `[Screenshot Placeholder: sync command output]`
- `[Screenshot Placeholder: target PostgreSQL tables]`
- `[Screenshot Placeholder: config example]`
