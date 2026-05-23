# ADR 0001 — Record architecture decisions

## Status
Accepted

## Context
The flagship will accumulate non-trivial design choices (cloud target, GitOps tool, secrets backend, registry). Without a written record, the rationale evaporates.

## Decision
Use lightweight ADRs in `docs/adr/`, numbered sequentially. Each ADR captures *Context → Decision → Consequences*. Past decisions are immutable; supersede with new ADRs.

## Consequences
- Anyone reading the repo can reconstruct *why*, not just *what*.
- Future-me (or interviewers) get a paper trail of engineering judgment.
