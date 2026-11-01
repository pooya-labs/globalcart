# GlobalCart backend foundation (M04)

This is the first runnable backend slice. It is deliberately separate from the Next.js static export hosted on GitHub Pages.

## Included
- Go 1.23 REST API using the standard HTTP mux.
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

## Local security boundary
- Compose publishes PostgreSQL and API only on loopback (127.0.0.1).
- The fallback password is for disposable local development only. Set `POSTGRES_PASSWORD` before using Compose and never reuse it in hosted environments.
- Do not expose this API to the public internet as-is.
- There is no authentication, user management, write API, payment, fulfillment, seller payout or production secret management yet.
- The catalog importer is a local/operator command, not an HTTP endpoint.
- GitHub Pages remains a static frontend; it does not run this service.

## Next M04 slices
1. Add integration tests against a disposable PostgreSQL service in CI.
2. Add sqlc generation and typed query layer after schema stabilizes.
3. Add identity/session and role/permission model before any admin/seller write endpoints.
4. Add deployment manifests and managed PostgreSQL only after the owner selects/approves a hosting provider, region, budget and secret-management approach.
5. Connect the frontend to the API only after a hosted HTTPS endpoint and CORS origins are approved.
