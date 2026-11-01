# Database migrations

Migration files are embedded into the API binary and applied in lexical order at startup. Each migration and its ledger entry run in one database transaction.

Rules:
- Never edit a migration that has reached a shared environment; add a new numbered migration.
- Keep schema changes backwards-compatible with the currently deployed API where practical.
- Use constraints and indexes to enforce invariants in PostgreSQL.
- Do not put secrets or real user data in migration fixtures.

- Number migrations with a stable, lexically sortable prefix; never rename an applied migration.
- Prefer additive changes and backfills that preserve existing API contracts.
- Use forward-fix migrations for shared environments. A destructive rollback is not automatic: assess data loss and coordinate a maintenance window before any reverse migration.
- Keep each migration compatible with the currently deployed application during rolling upgrades.


Migration 003 adds demo-only carts, checkout order snapshots, idempotency digests, temporary inventory reservations and order events. It is schema scaffolding only: do not expose write endpoints until the identity/deployment gates and the transaction protocol in `docs/COMMERCE_DESIGN.md` are satisfied. Never store raw cart capabilities or idempotency keys.
