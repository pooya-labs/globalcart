# Release & Handoff Process

## Versioning
- Patch: fixes and small refinements.
- Minor: a roadmap milestone or coherent feature slice.
- Major: stable v1.0 release or breaking architecture transition.

## Required updates
For every release update package.json, CHANGELOG.md, ROADMAP.md, and PROJECT_CONTEXT.md together.

## Verification
Run npm ci, npm run typecheck, npm run build, and relevant tests. Verify the GitHub Pages base path. Do not claim deployment success until Actions passes and the published URL is checked.

## Database recovery and migration release gate

Backend CI verifies the local recovery path on disposable PostgreSQL 17: after the race-enabled integration suite, it creates a custom-format logical dump, restores it into a fresh database, and compares the applied migration ledger, key table row counts, validated constraints and `pg_trgm` extension version. The temporary dump and restore database are removed by the job. This is a regression test, not a retained backup service or proof of production disaster recovery.

Before any shared/hosted database migration, require owner approval for the provider, region, encryption, retention, backup operator and recovery objectives (RPO/RTO). Take a provider-supported backup, verify a restore into an isolated database, validate the migration ledger and critical reads, then deploy the compatible application. Migrations are forward-only by default: prefer an additive forward-fix; restore a backup only when the write window and accepted data-loss boundary are understood. No hosted backup job or production migration automation is enabled by this repository.

## AI handoff
A new assistant must read PROJECT_CONTEXT.md, ROADMAP.md, CHANGELOG.md, docs/ARCHITECTURE.md, and docs/DECISIONS.md, then inspect current files and workflow status. The roadmap is the source of truth, not chat history.
