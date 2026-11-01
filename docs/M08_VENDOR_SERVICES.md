# M08.1 — Internal Vendor Services

This milestone adds internal-only Go service primitives on top of migration 005.

## Service methods

- `CreateVendor`: creates a merchant in `pending` status. This is not an account, membership or authorization grant.
- `UpsertVendorProduct`: creates or updates a base-product or variant offer after validating product ownership and requiring an active vendor. Initial/current price and stock audit entries are written in the same transaction.
- `UpdateVendorPrice`: locks the vendor and vendor offer, scopes the operation by both IDs, updates the price and appends a price-history row atomically.
- `AdjustVendorInventory`: locks the vendor and offer, rejects a negative resulting stock, updates stock and appends an inventory event atomically.
- `GetVendorProduct`: reads an offer scoped to the supplied vendor ID.

## Safety and compatibility

- Vendor creation defaults to `pending`; service mutations require `active`.
- Only USD is supported by this service until catalog currency/FX rules are designed.
- Audit records are transactionally coupled to the corresponding mutation.
- The current customer cart, reservation and checkout paths still use the legacy product/variant inventory model. This milestone does not claim multi-vendor checkout or split orders.
- There are no seller HTTP routes, seller memberships, seller dashboard, payouts, commission or fulfillment actions.
- Seller membership and tenant authorization remain blocked on the separately approved authentication and identity boundary.

## Verification

`vendors_integration_test.go` exercises PostgreSQL migrations, pending-vendor rejection, offer creation, price update, inventory restock, stock-underflow rejection, vendor-scoped readback and audit-row counts.
