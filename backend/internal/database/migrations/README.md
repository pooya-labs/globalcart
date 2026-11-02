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


Migration 012 adds nullable `carts.owner_user_id` referencing `users(id)`, plus a partial index for account-owned cart history. The application binds carts only after validating both the existing guest capability digest and a live customer session. Ownership is resolved from PostgreSQL in the same transaction as each account-backed cart/order operation; no HTTP route is enabled by this schema change.

## Migration integrity and recovery policy

The startup runner serializes migration work with a PostgreSQL advisory lock, applies each file and its ledger insert in one transaction, and stores the SHA-256 checksum of the exact embedded SQL bytes. It fails closed if an applied file changes or an applied ledger version has no matching embedded file. Legacy ledger rows with a NULL checksum are adopted once from the current canonical source tree; subsequent startup checks enforce the stored checksum. Never edit, rename or delete an applied migration, and never delete a ledger row to force a rerun.

Use a new numbered forward-fix migration for shared environments. Schema rollback is not automated. The current runner is transactional per file, so statements such as CREATE INDEX CONCURRENTLY are not supported inside embedded migrations. Read docs/DATABASE_MIGRATION_POLICY.md before creating, changing or recovering a migration.

Migration 013 adds administrator-owned product editorial overlays without mutating supplier fields. `admin_product_overrides` stores only title/description/brand overrides plus active/archived visibility, revision and actor metadata. `admin_product_revisions` stores before/after snapshots and has database-enforced append-only UPDATE/DELETE/TRUNCATE protection. Service mutations write the overlay, immutable revision and redacted audit event in one transaction. Catalog read queries apply the effective editorial fields and exclude archived products while preserving supplier source rows and dependent records. This migration does not expose admin HTTP routes or create manual products.

Migration 014 reserves a high-range sequence for administrator-created product IDs and assigns it as the product ID default while leaving explicitly supplied importer IDs unchanged. It also extends the append-only admin product revision action set with `created`. Manual product creation writes the product row, normalized category, platform demo offer, initial price/inventory events, override revision and audit event transactionally; its source is explicitly `admin_manual`.
