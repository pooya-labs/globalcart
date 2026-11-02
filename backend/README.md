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

## Identity and authorization schema (migration 011)

The local PostgreSQL schema now contains normalized users, migration-controlled customer/seller/admin roles, vendor-scoped seller memberships, server-side session digest records, single-use email-verification/password-reset token digests and append-only audit events. New accounts receive only the default customer role. The database rejects weak hash formats, malformed digest lengths, invalid expiry lifecycles, duplicate active grants/tokens and audit mutation.

The internal `backend/internal/auth` package provides Argon2id password hashing/verification with bounded PHC parsing, a rehash decision helper, 256-bit opaque-token generation/digesting, and a PostgreSQL-backed session lifecycle. Sessions are created only for active verified accounts, store only token digests, use a 15-minute idle timeout capped by a 24-hour absolute lifetime, resolve effective roles and active seller vendor scopes from PostgreSQL, validate session-bound CSRF tokens in constant time, and revoke/audit logout atomically.

The default Argon2id cost is 19 MiB, 2 iterations and 1 lane; benchmark this on the selected runtime and put rate limiting in front of expensive verification before enabling login. These internal functions are not wired to any handler. No login/register/session HTTP endpoint, browser cookie, email provider, live user collection or seller/admin route is enabled. The internal `RevokeUserSessions` helper revokes all active sessions in the caller's PostgreSQL transaction, locks the user row to serialize against session creation, accepts only migration-approved revocation reasons, and appends one redacted audit event per session. Call it in the same transaction as password, role or account-status changes; it is not an HTTP endpoint. ADR-010 remains proposed until the owner approves same-site HTTPS hosting, email delivery and secret management.

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
