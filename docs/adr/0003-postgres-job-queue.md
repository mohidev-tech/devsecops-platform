# ADR 0003 — Postgres as the job queue (Phase 1)

## Status
Accepted

## Context
Phase 1 needs a real workload to demonstrate the platform: api receives requests, worker processes them. The choice of queue mechanism affects deployment complexity, secret management surface, and what Phase 2's Vault story looks like.

Options considered:
1. **Postgres `FOR UPDATE SKIP LOCKED`** — single dependency, transactional, well-understood.
2. **Redis Streams / lists** — purpose-built, but adds a second stateful dependency.
3. **NATS / RabbitMQ** — even more deployment surface; overkill for portfolio scope.

## Decision
Use Postgres with `SELECT ... FOR UPDATE SKIP LOCKED` for the work queue. One StatefulSet for the database, both api and worker read DATABASE_URL from a single Kubernetes Secret.

## Consequences
- ✅ One stateful component, not two. Cleaner Phase 2 Vault story — one secret to migrate, not two.
- ✅ Transactional enqueue + claim — no message-lost vs. message-replayed tradeoff to explain.
- ✅ `SKIP LOCKED` lets worker replicas scale horizontally without external coordination.
- ⚠️ Postgres is not optimized for high-throughput queueing. Fine at portfolio scale (hundreds/sec); would not scale to a real production workload of millions/min.
- 📝 **Phase 1 ships plaintext dev creds in `values.yaml` deliberately** — the diff that replaces them with Vault-injected creds in Phase 2 is the literal proof of the security improvement. ADR 0004 (forthcoming) will record the swap.
