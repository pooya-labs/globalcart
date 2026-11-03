# GlobalCart backend foundation (M04)

This is the first runnable backend slice. It is deliberately separate from the Next.js static export hosted on GitHub Pages.

## Included
- Go REST API using the standard HTTP mux (module language baseline: Go 1.26.0; security-scanned CI and container build toolchain: Go 1.27.3).
- PostgreSQL 17 as the system of record.
- Embedded, versioned SQL migrations applied at API startup.\n- Catalog taxonomy with normalized category records and a backward-compatible product category link.\n- Product variants with constrained price/stock, optional unique SKU, source identity and JSONB attributes.
- Read-only catalog endpoints with bounded pagination and category/text filters.
- Liveness and database readiness probes.
- Explicit CORS origin allow-list.
- One-shot catalog snapshot importer for the existing normalized `data/catalog.json`.
- OpenAPI 3.1 contract.
- Docker Compose for local development only.

## Run locally

Requirements: Docker Engine with the Compose plugin.

From the repository root:

```sh
docker compose -f backend/docker-compose.yml up --build -d postgres api
docker compose -f backend/docker-compose.yml --profile tools run --rm catalog-import
```

Then check:

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
curl 'http://localhost:8080/api/v1/products?limit=10'
curl http://localhost:8080/api/v1/products/1
```

Stop services:

```sh
docker compose -f backend/docker-compose.yml down
```

Keep the database volume by default. To delete local development data permanently, use `docker compose -f backend/docker-compose.yml down -v`.

## Runtime configuration

The API reads these environment variables at startup:

| Variable | Default / requirement | Validation |
|---|---|---|
| `APP_ENV` | `development` when omitted | Only `development`, `test` and `production` are accepted. |
| `DATABASE_URL` | Required | Must be a PostgreSQL URL with a host and database name. Production must use `sslmode=require`, `verify-ca` or `verify-full`. |
| `HTTP_ADDR` | `:8080` | Must be a valid `host:port` TCP listener with a port from 1 to 65535. |
| `CORS_ORIGINS` | Empty allow-list | Comma-separated exact origins only. Wildcard and `null` are rejected; paths, credentials, queries and fragments are not accepted. Plain HTTP is limited to loopback origins in development/test; production requires HTTPS. |

Local Compose sets `APP_ENV=development` and allows `http://localhost:3000`. For a future production deployment, set `APP_ENV=production`, use a TLS-capable PostgreSQL URL and list the exact HTTPS frontend origin. This validation does not terminate TLS, provision infrastructure or create/manage secrets; those remain deployment prerequisites.

Example production-shaped values (placeholders only):

```dotenv
APP_ENV=production
DATABASE_URL=postgres://globalcart:replace-me@db.example:5432/globalcart?sslmode=verify-full
HTTP_ADDR=:8080
CORS_ORIGINS=https://shop.example.com
```

## Identity and authorization schema (migration 011)

The local PostgreSQL schema now contains normalized users, migration-controlled customer/seller/admin roles, vendor-scoped seller memberships, server-side session digest records, single-use email-verification/password-reset token digests and append-only audit events. New accounts receive only the default customer role. The internal email-verification lifecycle reuses migration 011, stores only token digests, rotates previous tokens and activates only pending accounts; it preserves the customer role and cannot grant seller/admin access. Issuance and consumption have PostgreSQL race-enabled test coverage. Internal `GrantAdminRole` and `RevokeAdminRole` require a different active, verified admin actor, serialize role changes, revoke target sessions and append audit events atomically. The trusted caller must derive actor identity from a verified session; no HTTP role endpoint or bootstrap API exists. The pure `Authorize` policy evaluator accepts only a `SessionPrincipal` from `ResolveSession`, denies unknown actions and checks customer ownership, seller vendor scope and explicit admin actions. It does not replace canonical ownership/vendor checks in the protected service transaction and is not wired to HTTP. Seller-specific vendor read/price/stock methods now use `AuthorizeSellerVendorTx` to lock and revalidate the account, exact session, active vendor, seller role and active membership in the same PostgreSQL transaction as the operation. These methods derive vendor scope from the canonical offer record; cross-vendor IDOR and membership-revocation race tests pass. They remain internal and are not connected to HTTP. Internal seller-membership request/activation/suspension/revocation reuses migration 011, grants the seller role only after admin approval, preserves other active vendor scopes, removes seller role after the final scope is gone and revokes sessions/audits transactionally. No seller HTTP route or dashboard is enabled. The database rejects weak hash formats, malformed digest lengths, invalid expiry lifecycles, duplicate active grants/tokens and audit mutation.

The internal `backend/internal/auth` package provides Argon2id password hashing/verification with bounded PHC parsing, a rehash decision helper, 256-bit opaque-token generation/digesting, and a PostgreSQL-backed session lifecycle. Sessions are created only for active verified accounts, store only token digests, use a 15-minute idle timeout capped by a 24-hour absolute lifetime, resolve effective roles and active seller vendor scopes from PostgreSQL, validate session-bound CSRF tokens in constant time, and revoke/audit logout atomically.

The default Argon2id cost is 19 MiB, 2 iterations and 1 lane; benchmark this on the selected runtime and put rate limiting in front of expensive verification before enabling login. These internal functions are not wired to any handler. No login/register/session HTTP endpoint, browser cookie, email provider, live user collection or seller/admin route is enabled. The internal `RevokeUserSessions` helper revokes all active sessions in the caller's PostgreSQL transaction, locks the user row to serialize against session creation, accepts only migration-approved revocation reasons, and appends one redacted audit event per session. Call it in the same transaction as password, role or account-status changes; it is not an HTTP endpoint. `CreateSession` also requires the exact password-hash snapshot verified by the caller and rejects it if a concurrent password change has replaced it. Internal `ChangePassword` updates the hash, revokes sessions and appends a password-change audit event atomically. These functions are not HTTP endpoints. Internal password-reset token issuance/consumption reuses migration 011, stores only token digests, rotates older tokens and completes a reset atomically with sibling-token/session revocation and audit. Raw reset tokens are returned once to an internal caller; no email provider or public route is connected. PostgreSQL race-enabled Backend CI and repository security scans passed for this internal lifecycle. The PostgreSQL race-enabled integration suite verifies stale login rejection and atomic password-change/session revocation. ADR-010 remains proposed until the owner approves same-site HTTPS hosting, email delivery and secret management.

## Customer ownership boundary (M04.12; internal only)

Migration 012 adds nullable `carts.owner_user_id`. `BindCartToCustomer` requires both the guest cart capability digest and a live verified customer session. Binding is serialized by the cart row lock; the first account wins. The internal `backend/internal/commerce/customer_ownership.go` service provides account-backed cart read/add/set/remove, parent and vendor-child order reads, and unpaid simulated-order cancellation. Each operation revalidates the customer session/role and loads canonical ownership from PostgreSQL in the same transaction; owner IDs from request data are never accepted. Foreign and guest-only resources return non-enumerating not-found errors.

Guest cart capabilities remain independent: possession authorizes only that one cart and does not grant an account identity or any role. The account-backed methods remain internal and are not registered on the HTTP router. PostgreSQL integration tests cover cross-account IDOR, cart mutation, child-order reads, binding races, expired carts, revoked sessions and cancellation replay. Backend CI [41403135351](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135351), Security scans [41403135364](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135364) and Pages deployment [41403135344](https://github.com/pooya16pg-spec/globalcart/actions/runs/41403135344) passed on implementation head `928c3867e2d8e109e4e6595f9ac0898cf51b25b9`.

## Local security boundary
- Compose publishes PostgreSQL and API only on loopback (127.0.0.1).
- The fallback password is for disposable local development only. Set `POSTGRES_PASSWORD` before using Compose and never reuse it in hosted environments.
- Do not expose this API to the public internet as-is.
- There is no public authentication/session HTTP flow, user-management UI, write API, payment, fulfillment, seller payout or production secret management yet.
- The catalog importer is a local/operator command, not an HTTP endpoint.
- GitHub Pages remains a static frontend; it does not run this service.

## Remaining M04 gates
1. Complete authentication/session/RBAC middleware and authorization tests on top of migration 011, only after same-site HTTPS hosting, email delivery and secret storage are approved.
2. Approve the hosting provider, region, budget and secret-management approach before provisioning a managed PostgreSQL instance or public API.
3. Approve a custom frontend domain and same-site HTTPS API origin before connecting the static storefront to the backend.
4. Add deployment manifests, production secret handling, backup/restore and smoke tests only after the hosting decisions are approved.

The disposable-PostgreSQL CI suite, generated sqlc layer, read-only handlers and Docker image build are already implemented and verified; they are not pending M04 tasks.


## Typed SQL layer

The root `sqlc.yaml` defines generated pgx/v5 query code from `backend/queries` and the embedded migration schema. Run:

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
```

CI regenerates and formats the typed package before tests/builds. The read-only API handlers already use generated typed queries for product listing/counting, product detail, categories and active product variants.


The public product list and detail read handlers use the generated sqlc query package. The importer intentionally retains explicit transactional upsert statements because it performs a multi-step product/category transaction rather than a simple generated read query.


The read-only catalog API also exposes `GET /api/v1/products/{id}/variants` for active product variants. It returns a 404 when the parent product does not exist.


## Read-only catalog filters

`GET /api/v1/products` accepts:
- `q`: bounded full-text query.
- `category`: case/whitespace-normalized category name.
- `brand`: case/whitespace-normalized exact brand.
- `minPrice` / `maxPrice`: inclusive source-price bounds, non-negative with at most two decimal places.
- `minRating` / `maxRating`: inclusive rating bounds from 0 to 5, with at most two decimal places.
- `limit` / `offset`: bounded pagination.

The paginated rows and `total` count use the same predicates. Invalid numeric input, precision, range or inverted bounds returns HTTP 400 before the database is queried. Migration 008 adds supporting brand, price and rating indexes. PostgreSQL integration tests cover combined filters and count/page consistency.

The storefront currently applies the corresponding brand and minimum-rating facets locally against the generated catalog snapshot. It is not wired to this local-only API; that integration remains gated on approved hosting and same-site HTTPS.


## Internal demo return/refund simulation (M05.1)
Migration 010 and `backend/internal/commerce/returns.go` support one full-order return request for a paid simulated order, fixed reason codes, internal approve/reject review and owner-scoped simulated refund completion. The refund amount is the persisted net order total, including coupon discounts. Parent and vendor child order states are updated transactionally and an append-only order event is recorded. Integration tests are in `returns_integration_test.go`.

These methods are not registered with the HTTP router. The review operation has no authenticated reviewer identity yet, so it must remain internal until identity/session/RBAC, audit policy and hosting are approved. No real payment provider, movement of funds, shipment or inventory adjustment occurs.

## Migration integrity and recovery

The migration ledger stores the applied filename, timestamp and SHA-256 checksum of the exact SQL file bytes. Startup fails closed if an applied migration is modified or its source file disappears. Existing pre-checksum ledger rows are backfilled once; new rows require a checksum. Each file and ledger insert commit atomically, and the runner serializes concurrent startup attempts with a PostgreSQL advisory lock.

Applied migrations are immutable. Use a new numbered forward-fix migration; do not delete ledger entries or assume schema rollback is safe. The current runner executes each migration in a transaction, so non-transactional statements such as CREATE INDEX CONCURRENTLY require a separately designed operational procedure. See docs/DATABASE_MIGRATION_POLICY.md for expand-and-contract, backup/restore, rollback decisions and the incident runbook.

## Internal admin overview foundation (M06.1; not an HTTP endpoint)

`backend/internal/admin.ReadOverview(ctx, pool, sessionToken)` returns bounded aggregate diagnostics for the local PostgreSQL platform only after resolving the opaque server-side session token and requiring an active administrator role. It reports catalog counts/quality/source breakdown, variant and vendor-offer states, seller-membership/account status totals, price-history coverage and recency, audit event volume/recency, and the intentionally paused translation state.

The service revalidates the account, session and active admin-role rows under locks in the same transaction as the overview query. A stale in-memory role/session principal is not sufficient. The response intentionally excludes emails, password hashes, raw or hashed session/CSRF tokens, and audit-event metadata/payloads.

The service is internal-only and is not registered on `backend/internal/httpapi`. No dashboard, public admin route, role switcher, user bootstrap, account mutation or seller write operation is enabled. Do not expose it until the approved same-site HTTPS identity middleware and hosting/email/secret prerequisites are complete. The test suite exercises PostgreSQL-backed customer/seller denial, administrator access, account/session/role revocation and aggregate data boundaries.

## Internal admin product management (M06.2; not HTTP endpoints)

Migrations 013–014 add administrator editorial overrides, immutable revision history and a reserved high-range product ID sequence for manual products. The internal `backend/internal/admin` service supports administrator-only manual product creation, managed reads, bounded title/description/brand overrides, explicit reset-to-current-supplier values, archive/restore and bounded revision-history reads.

Manual products use `source=admin_manual` and a generated source ID. Creation inserts the catalog row, category, active GlobalCart Direct demo offer, opening price/stock history, revision 1 and a redacted audit event in one transaction. The reserved sequence avoids collisions with current supplier IDs and is assigned as the default product ID for future inserts that omit an ID.

Edits never write to supplier-owned product fields. Source imports continue updating their own record, while catalog reads apply editorial overrides to effective title, description and brand. Archived products are filtered from list/detail, search, totals and category counts without deleting source rows, variants or vendor offers.

Every edit/archive/restore call must provide the expected current revision. Stale mutations fail with `ErrStaleProductRevision`. The service resolves the opaque session token, requires `admin.products.manage`, then revalidates the verified account, exact live session and admin grant under row locks in the same transaction as the protected read/write. Before/after snapshots retain effective and supplier values; revision history is append-only and a redacted audit event is written atomically.

These remain internal services only. No admin HTTP route, dashboard, browser role switcher or public catalog write path is enabled. The storefront is not connected to these backend services from GitHub Pages.

## Internal administrator product work queue (M06.3; not an HTTP endpoint)

`backend/internal/admin.ListManagedProducts` provides a bounded list for the future administrator workspace. It requires an opaque live session and the existing `admin.products.manage` permission; the account, exact session and admin grant are revalidated under row locks in the protected transaction.

The query supports effective title/description/brand search, category, source, source-ID and product-ID matching, active (default), archived or all visibility states, exact source filtering, effective-brand filtering, and a fixed sort allow-list (updated time, import time, title, price, rating or stock). Pagination defaults to 20 rows, allows at most 100 rows and caps the offset at 10,000. A total count is returned with each page. User-provided sort strings are never interpolated into SQL.

This is an internal service only. It adds no migration, HTTP handler, public endpoint, browser dashboard or GitHub Pages integration. Supplier provenance is read from the original catalog row; editorial search uses the separate effective override values. PostgreSQL integration tests cover non-admin denial, filters, archive visibility, ordering, pagination and invalid bounds.
