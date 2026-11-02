# PostgreSQL Migration, Rollback and Recovery Policy

**Status:** Adopted for repository development and future shared environments  
**Scope:** Go/PostgreSQL migrations embedded by `backend/internal/database`  
**Current deployment boundary:** PostgreSQL is exercised in local development and disposable CI. No hosted production database or public write API is approved.

This policy is operational guidance, not authorization to provision infrastructure, run destructive SQL, or expose backend routes.

## 1. Current migration-runner guarantees

The current `Migrate` implementation:

- discovers embedded `migrations/*.sql` files and applies them in lexical filename order;
- takes PostgreSQL advisory lock `7142032601` so concurrent API/importer startup attempts serialize;
- executes each individual migration and its ledger insert in one PostgreSQL transaction;
- records the filename, application timestamp and SHA-256 checksum of the exact embedded SQL bytes in `schema_migrations`;
- rejects a changed applied file and rejects a ledger version whose embedded migration file is missing;
- adopts a NULL checksum from a pre-checksum ledger once, then makes the checksum column NOT NULL.

A failed migration transaction is rolled back and does not receive a ledger entry. A later migration failure does not undo earlier migrations that already committed: the transaction boundary is one file, not the entire migration directory.

The runner splits SQL and executes statements individually inside that transaction. Statements that PostgreSQL forbids inside a transaction (for example, `CREATE INDEX CONCURRENTLY`) are not supported by this runner. Do not add them to an embedded migration; large-table online index work needs a separately designed and reviewed operational path.

## 2. Immutable migration rules

1. Use a new zero-padded, lexically sortable filename, for example `013_add_order_search_index.sql`.
2. Treat every migration merged to `main` as immutable. Never edit, rename, reorder or delete an applied migration.
3. Do not delete a `schema_migrations` row to make a migration run again. Do not manually mark a migration applied without verifying the exact schema and data state.
4. Keep each migration narrowly scoped, deterministic and safe to retry only through the runner's ledger behavior. Do not put secrets, production data or environment-specific values in SQL fixtures.
5. Prefer additive schema changes. Keep old and new application versions compatible during a rolling deployment.
6. Add database constraints for invariants, plus tests for both valid and invalid data. Large backfills must be bounded, observable and idempotent.
7. Review migration SQL, its expected lock duration, table size, indexes, rewrite behavior, rollback options and data-retention impact before applying it to shared data.

The checksum is a guard against accidental or unauthorized migration-file drift, not a substitute for code review, signed releases or database access control. The one-time NULL-checksum adoption records the repository's current canonical bytes for databases created before checksum tracking; it cannot reconstruct historical file contents that were changed before this guard existed.

## 3. Expand-and-contract deployment sequence

For a future shared environment, use separate compatible releases rather than combining schema removal and application cutover:

1. **Preflight:** identify the target migration range and current ledger; review SQL and expected locks; verify available disk, PostgreSQL version and extension prerequisites.
2. **Backup:** take an environment-consistent backup using the approved provider's backup/PITR facility and, where appropriate, a custom-format logical dump. Record its timestamp and target database identity.
3. **Restore rehearsal:** restore the backup into an isolated database and verify the migration ledger, critical row counts, constraints and application-level invariants. A successful dump command alone is not a restore test.
4. **Expand:** add nullable columns, new tables, compatible indexes and dual-read/dual-write support first. Avoid renaming or dropping fields still used by the currently deployed application.
5. **Backfill:** run bounded, restartable and idempotent batches. Track progress and verify counts/checksums before switching reads.
6. **Switch:** deploy application code that uses the new representation while remaining compatible with the expanded schema. Observe errors, readiness and domain-specific invariants.
7. **Contract later:** only after all old application instances/jobs are gone, a separate reviewed migration may remove obsolete fields or constraints. Take a fresh backup first and prove no supported code path still uses them.

The current runner applies migrations at startup. Its advisory lock prevents two cooperating instances from applying the same migration concurrently, but it does not remove the need to control which release is allowed to perform DDL in a shared environment. A future multi-replica hosted deployment should use a single, explicitly approved migration job/release step or otherwise tightly control startup DDL privileges.

## 4. Rollback decision policy

**Default: forward-fix the database; roll back application code only when compatibility is proven.**

- **Application defect, compatible schema:** stop rollout and roll back to the last known-good application artifact. Keep the additive schema in place.
- **Migration fails before commit:** the current file transaction and ledger insert roll back together. Diagnose the failure. If no shared environment could have applied the file, correct it before release; once any shared environment may have applied it, preserve the file and use a new forward-fix migration.
- **Migration committed but schema/data is wrong:** stop the rollout, preserve a snapshot for diagnosis and write a new, reviewed forward-fix migration. Do not edit the applied SQL or erase its ledger row.
- **Destructive data transformation:** design a compensating migration only when the previous values were deliberately retained and can be restored without guessing. Otherwise restore a verified backup into an isolated database, validate it, and plan a controlled cutover. Restoring an old backup over a live database can discard legitimate writes made after the backup.
- **Irrecoverable corruption or suspected compromise:** pause writes, preserve forensic evidence, involve the database/operator owner and follow the provider's recovery procedure. Do not improvise by dropping tables or replaying migrations manually.

There are no automatic `down.sql` files or automatic destructive rollback commands. Any exception requires an explicit, reviewed operational plan and owner approval for the affected shared environment.

## 5. Pre-merge checklist

- [ ] A new migration filename is used; no applied migration was edited or removed.
- [ ] The SQL is transactional under the current runner and contains no unsupported non-transactional statement.
- [ ] The change follows expand-and-contract where old/new code overlap is possible.
- [ ] Locking, rewrite, index build and backfill costs are understood for the target data size.
- [ ] Constraints, indexes, foreign keys and database numeric bounds are covered by PostgreSQL integration tests.
- [ ] Failure behavior is tested: the migration transaction and ledger row either both commit or neither commits.
- [ ] A restore path and data-loss trade-off are documented for any destructive or irreversible operation.
- [ ] Backend CI, race-enabled PostgreSQL tests, migration integrity tests and security scans pass.

## 6. Incident runbook

1. Stop the deployment or affected write path; do not immediately rerun a failed release.
2. Capture the application commit, migration filenames/checksums, PostgreSQL version and the complete error. Preserve the current database state for diagnosis.
3. Check `schema_migrations` and the actual schema/data to determine whether the failing file committed. Never infer success from process logs alone.
4. Take or preserve a recovery point before further mutation. Restore it into an isolated database and validate it.
5. Choose the least destructive recovery: application rollback when schema-compatible, otherwise a reviewed forward-fix. Use backup cutover only with a write-free/reconciliation plan.
6. Verify the ledger, constraints, critical counts and domain invariants, then run health/readiness and application smoke tests before reopening traffic.
7. Record the incident, recovery point, actions, approvals and remaining risks in the project handoff.

## 7. Backup and recovery boundary today

Backend CI creates a PostgreSQL 17 custom-format logical dump, restores it into a fresh disposable database with `pg_restore --exit-on-error`, and compares the ordered migration ledger, selected table counts, validated constraint count and `pg_trgm` version.

That test proves the CI schema can be dumped and restored. It does **not** provide an off-site or encrypted production backup, point-in-time recovery, retention policy, production RPO/RTO, alerting or a provider-backed recovery drill. Those require an approved hosting provider, region, budget, secret-management and backup policy before a public database is provisioned.
