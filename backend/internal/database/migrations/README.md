# Database migrations

Migration files are embedded into the API binary and applied in lexical order at startup. Each migration and its ledger entry run in one database transaction.

Rules:
- Never edit a migration that has reached a shared environment; add a new numbered migration.
- Keep schema changes backwards-compatible with the currently deployed API where practical.
- Use constraints and indexes to enforce invariants in PostgreSQL.
- Do not put secrets or real user data in migration fixtures.
