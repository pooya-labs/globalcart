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

Migration 007 adds the missing PostgreSQL upper bound for product ratings (`rating <= 5`). This matches the canonical catalog domain and the snapshot validator; keep the application and database limits aligned.


Migration 009 enables `pg_trgm` and adds GIN indexes for normalized product titles and brands. Catalog search keeps PostgreSQL full-text matching as its primary path, then uses trigram/word similarity as a bounded fallback for queries of at least three characters. Exact full-text matches rank ahead of fuzzy matches. Keep both the list and count query predicates aligned.


Migration 010 adds a full-order demo return record, fixed reason codes and lifecycle constraints. Return requests are limited to paid simulated parent orders; approval records the persisted net order total and final completion changes parent/vendor child status to `refunded_simulated`. This is an internal state machine only: no payment provider or inventory mutation is involved, and no public HTTP route is enabled.


Migration 011 adds the identity/session/RBAC schema foundation: normalized users, a migration-seeded role catalog, seller memberships, SHA-256 session/CSRF and action-token digests, and append-only audit events enforced by PostgreSQL triggers. New users receive only the customer role. This is storage-only; do not infer that authentication or seller authorization is available. The migration runner understands PostgreSQL dollar-quoted function bodies as well as ordinary quoted strings and comments.
