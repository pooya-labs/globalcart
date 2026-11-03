# GlobalCart — PostgreSQL Entity Relationship Diagram

> **Schema source of truth:** the append-only SQL migrations under `backend/internal/database/migrations/`.
> This document reflects migrations **001–015** on `main`; it does not describe a deployed production database.

## Scope and verification

The current schema defines **27 tables across 15 versioned migrations**. The diagrams below are split by domain so the relationships remain readable. They include the foreign-key relationships declared by the migrations; application-level behavior and browser-local state are described separately.

| Domain | Tables | Main responsibility |
|---|---:|---|
| Catalog and marketplace | 8 | Source products, categories, variants, vendors, offers and offer history |
| Demo commerce | 10 | Carts, parent/vendor orders, immutable item snapshots, reservations, coupons, returns and events |
| Identity and access | 7 | Accounts, roles, seller memberships, sessions, one-time tokens and audit events |
| Editorial and offer history | 3 | Admin product overrides/revisions and seller offer revisions |
| **Total** | **27** | |

The counts are table ownership groupings for documentation; shared entities such as `products`, `vendors` and `users` appear as references in more than one diagram but are counted once.

## 1. Catalog and multi-vendor marketplace

```mermaid
erDiagram
    CATEGORIES {
        bigint id PK
        text name
        bigint parent_id FK "nullable self-reference"
        timestamptz created_at
    }
    PRODUCTS {
        bigint id PK
        text title
        text description
        text category
        bigint category_id FK "nullable"
        numeric price
        numeric discount_percentage
        numeric rating
        integer stock
        text brand
        text source
        text source_id
        timestamptz imported_at
    }
    PRODUCT_VARIANTS {
        bigint id PK
        bigint product_id FK
        text sku
        text title
        numeric price
        integer stock
        jsonb attributes
        text source
        text source_id
        boolean is_active
    }
    VENDORS {
        bigint id PK
        text slug UK
        text display_name
        text vendor_type
        text status
    }
    VENDOR_PRODUCTS {
        bigint id PK
        bigint vendor_id FK
        bigint product_id FK
        bigint variant_id FK "nullable; paired with product_id"
        text seller_sku
        numeric price
        char currency
        integer stock
        text status
        bigint revision
    }
    VENDOR_PRICE_HISTORY {
        bigint id PK
        bigint vendor_product_id FK
        numeric old_price
        numeric new_price
        char currency
        text change_source
        timestamptz changed_at
    }
    VENDOR_INVENTORY_EVENTS {
        bigint id PK
        bigint vendor_product_id FK
        integer quantity_delta
        integer resulting_stock
        text reason
        text reference
        timestamptz created_at
    }

    CATEGORIES o|--o{ CATEGORIES : "parent / child"
    CATEGORIES o|--o{ PRODUCTS : categorizes
    PRODUCTS ||--o{ PRODUCT_VARIANTS : has
    VENDORS ||--o{ VENDOR_PRODUCTS : lists
    PRODUCTS ||--o{ VENDOR_PRODUCTS : offered_as
    PRODUCT_VARIANTS o|--o{ VENDOR_PRODUCTS : variant_offer
    VENDOR_PRODUCTS ||--o{ VENDOR_PRICE_HISTORY : price_changes
    VENDOR_PRODUCTS ||--o{ VENDOR_INVENTORY_EVENTS : stock_events
```

**Important constraints**

- A product's normalized `category_id` is optional; the original source `category` text remains on `products`.
- A variant belongs to exactly one product. Variant references from offers use the composite `(variant_id, product_id)` key so a variant cannot be attached to a different product.
- A vendor can have one base offer per product or one offer per variant. Partial unique indexes enforce those two cases independently.
- `vendor_products` owns seller price, currency, stock and lifecycle status. Imported product and variant records retain their source provenance.
- `vendor_price_history` and `vendor_inventory_events` are append-oriented histories linked to the offer; they are not alternative inventory or price sources.

## 2. Demo commerce, coupons and order lifecycle

```mermaid
erDiagram
    USERS {
        uuid id PK
    }
    COUPONS {
        bigint id PK
        bytea code_hash UK
        text discount_type
        numeric discount_value
        numeric minimum_subtotal
        numeric maximum_discount
        timestamptz starts_at
        timestamptz ends_at
        bigint usage_limit
        integer per_customer_limit
        boolean is_active
    }
    CARTS {
        uuid id PK
        bytea owner_token_hash UK
        uuid owner_user_id FK "nullable"
        bigint coupon_id FK "nullable"
        text status
        char currency
        timestamptz expires_at
    }
    CART_ITEMS {
        bigint id PK
        uuid cart_id FK
        bigint product_id FK
        bigint variant_id FK "nullable"
        bigint vendor_product_id FK "nullable"
        integer quantity
        numeric unit_price_snapshot
        char currency
        text product_title_snapshot
    }
    DEMO_ORDERS {
        uuid id PK
        uuid cart_id FK
        bigint coupon_id FK "nullable"
        bytea idempotency_key_hash
        text status
        numeric subtotal
        numeric discount_amount
        numeric total
        char currency
    }
    DEMO_VENDOR_ORDERS {
        uuid id PK
        uuid order_id FK
        bigint vendor_id FK
        text status
        numeric subtotal
        numeric discount_amount
        numeric total
        char currency
    }
    DEMO_ORDER_ITEMS {
        bigint id PK
        uuid order_id FK
        uuid vendor_order_id FK "nullable"
        bigint product_id FK
        bigint variant_id FK "nullable"
        bigint vendor_product_id FK "nullable"
        integer quantity
        numeric unit_price_snapshot
        char currency
        text product_title_snapshot
    }
    INVENTORY_RESERVATIONS {
        uuid id PK
        uuid order_id FK
        uuid vendor_order_id FK "nullable"
        bigint product_id FK
        bigint variant_id FK "nullable"
        bigint vendor_product_id FK "nullable"
        integer quantity
        text status
        timestamptz expires_at
    }
    DEMO_ORDER_EVENTS {
        bigint id PK
        uuid order_id FK
        text event_type
        jsonb payload
        timestamptz created_at
    }
    COUPON_REDEMPTIONS {
        bigint id PK
        bigint coupon_id FK
        uuid order_id FK "unique"
        bytea owner_token_hash
        numeric discount_amount
        timestamptz redeemed_at
    }
    DEMO_RETURNS {
        uuid id PK
        uuid order_id FK "unique"
        text reason
        text status
        char currency
        numeric refund_amount
        timestamptz reviewed_at
        timestamptz refunded_at
    }

    USERS o|--o{ CARTS : owns
    COUPONS o|--o{ CARTS : applied_to
    CARTS ||--o{ CART_ITEMS : contains
    PRODUCTS ||--o{ CART_ITEMS : selected
    PRODUCT_VARIANTS o|--o{ CART_ITEMS : variant
    VENDOR_PRODUCTS o|--o{ CART_ITEMS : seller_offer
    CARTS ||--o{ DEMO_ORDERS : checked_out_as
    COUPONS o|--o{ DEMO_ORDERS : discounts
    DEMO_ORDERS ||--o{ DEMO_VENDOR_ORDERS : split_into
    VENDORS ||--o{ DEMO_VENDOR_ORDERS : fulfills_in_demo
    DEMO_ORDERS ||--o{ DEMO_ORDER_ITEMS : snapshots
    DEMO_VENDOR_ORDERS o|--o{ DEMO_ORDER_ITEMS : vendor_lines
    PRODUCTS ||--o{ DEMO_ORDER_ITEMS : ordered
    PRODUCT_VARIANTS o|--o{ DEMO_ORDER_ITEMS : variant
    VENDOR_PRODUCTS o|--o{ DEMO_ORDER_ITEMS : offer_snapshot
    DEMO_ORDERS ||--o{ INVENTORY_RESERVATIONS : reserves
    DEMO_VENDOR_ORDERS o|--o{ INVENTORY_RESERVATIONS : vendor_reservation
    PRODUCTS ||--o{ INVENTORY_RESERVATIONS : product_stock
    PRODUCT_VARIANTS o|--o{ INVENTORY_RESERVATIONS : variant_stock
    VENDOR_PRODUCTS o|--o{ INVENTORY_RESERVATIONS : offer_stock
    DEMO_ORDERS ||--o{ DEMO_ORDER_EVENTS : records
    COUPONS ||--o{ COUPON_REDEMPTIONS : redeemed
    DEMO_ORDERS ||--o| COUPON_REDEMPTIONS : redemption
    DEMO_ORDERS ||--o| DEMO_RETURNS : return_request
```

**Important constraints**

- Carts may be guest-owned through a digest-only `owner_token_hash` or associated with a verified account through nullable `owner_user_id`. Orders do not have a direct `user_id` column; account ownership is resolved through the cart.
- `cart_items` and order/reservation rows support both legacy product/variant lines and vendor-offer lines. Vendor-order and vendor-offer references must be supplied together on order items and reservations.
- Parent orders can split into one child order per vendor. Composite foreign keys ensure each child line/reservation belongs to the same parent order as its vendor child order.
- Product titles and unit prices are snapshots on cart/order lines; those fields preserve the values presented at the time of the demo flow.
- Coupon codes and guest capabilities are stored as hashes, not reusable plaintext secrets.
- `demo_returns` permits at most one full-order return per parent order. Refund outcomes are simulated state changes only.

## 3. Identity, sessions and authorization

```mermaid
erDiagram
    ROLES {
        text role_key PK
        text display_name
    }
    USERS {
        uuid id PK
        text email_normalized UK
        text password_hash
        text status
        timestamptz email_verified_at
        timestamptz disabled_at
    }
    USER_ROLES {
        bigint id PK
        uuid user_id FK
        text role_key FK
        uuid granted_by_user_id FK "nullable"
        timestamptz granted_at
        timestamptz revoked_at
    }
    SELLER_MEMBERSHIPS {
        uuid id PK
        uuid user_id FK
        bigint vendor_id FK
        text status
        uuid reviewed_by_user_id FK "nullable"
        timestamptz activated_at
    }
    USER_SESSIONS {
        uuid id PK
        uuid user_id FK
        bytea token_digest UK
        bytea csrf_token_digest
        timestamptz idle_expires_at
        timestamptz expires_at
        timestamptz revoked_at
    }
    AUTH_ACTION_TOKENS {
        uuid id PK
        uuid user_id FK
        text purpose
        bytea token_digest UK
        timestamptz expires_at
        timestamptz consumed_at
        timestamptz revoked_at
    }
    AUDIT_EVENTS {
        bigint id PK
        text event_type
        uuid actor_user_id FK "nullable"
        uuid subject_user_id FK "nullable"
        text resource_type
        text resource_id
        text request_id
        jsonb metadata
        timestamptz created_at
    }
    VENDORS {
        bigint id PK
    }

    USERS ||--o{ USER_ROLES : assigned
    ROLES ||--o{ USER_ROLES : grants
    USERS o|--o{ USER_ROLES : granted_by
    USERS ||--o{ SELLER_MEMBERSHIPS : requests
    VENDORS ||--o{ SELLER_MEMBERSHIPS : has_members
    USERS o|--o{ SELLER_MEMBERSHIPS : reviewed_by
    USERS ||--o{ USER_SESSIONS : opens
    USERS ||--o{ AUTH_ACTION_TOKENS : receives
    USERS o|--o{ AUDIT_EVENTS : actor
    USERS o|--o{ AUDIT_EVENTS : subject
```

**Important constraints**

- Every inserted account receives the `customer` role from a database trigger. Seller and admin grants are separate, audited operations.
- One active grant per user/role and one membership per user/vendor are enforced by unique constraints/indexes.
- Sessions and one-time action tokens persist SHA-256 digests, not raw bearer tokens. Token purposes are limited to email verification and password reset.
- Active seller access is scoped through a vendor membership; membership status is distinct from the user's global role.
- Audit events are append-only at the database boundary. Their metadata is bounded and rejects keys that could carry passwords, tokens, cookies, secrets or card numbers.

## 4. Editorial administration and revision history

```mermaid
erDiagram
    PRODUCTS {
        bigint id PK
        text source
        text source_id
    }
    USERS {
        uuid id PK
    }
    VENDORS {
        bigint id PK
    }
    VENDOR_PRODUCTS {
        bigint id PK
        bigint vendor_id FK
        bigint product_id FK
        bigint revision
    }
    ADMIN_PRODUCT_OVERRIDES {
        bigint product_id PK, FK
        text title_override
        text description_override
        text brand_override
        text visibility_status
        bigint revision
        uuid created_by_user_id FK
        uuid updated_by_user_id FK
        uuid archived_by_user_id FK "nullable"
        timestamptz archived_at
    }
    ADMIN_PRODUCT_REVISIONS {
        bigint id PK
        bigint product_id FK
        bigint revision
        text action
        jsonb before_state
        jsonb after_state
        uuid actor_user_id FK
    }
    VENDOR_PRODUCT_REVISIONS {
        bigint id PK
        bigint vendor_product_id FK
        bigint revision
        text action
        jsonb before_state
        jsonb after_state
        uuid actor_user_id FK
    }

    PRODUCTS ||--o| ADMIN_PRODUCT_OVERRIDES : editorial_override
    PRODUCTS ||--o{ ADMIN_PRODUCT_REVISIONS : revision_log
    USERS ||--o{ ADMIN_PRODUCT_OVERRIDES : created_by
    USERS ||--o{ ADMIN_PRODUCT_OVERRIDES : updated_by
    USERS o|--o{ ADMIN_PRODUCT_OVERRIDES : archived_by
    USERS ||--o{ ADMIN_PRODUCT_REVISIONS : actor
    VENDOR_PRODUCTS ||--o{ VENDOR_PRODUCT_REVISIONS : revision_log
    USERS ||--o{ VENDOR_PRODUCT_REVISIONS : actor
```

**Important constraints**

- Supplier-owned product fields are not overwritten by editorial edits. Admin overrides are stored separately and take precedence only in the effective catalog view; clearing an override returns to the current supplier value.
- Admin revision history and seller offer revision history are append-only. Each revision is tied to its product/offer and actor.
- Seller offer updates advance the offer's optimistic-concurrency revision in a database trigger. Admin edits/archive/restore use expected revisions in the service layer.
- Manual admin-created product IDs use the reserved high range configured by migration 014; the sequence is owned by `products.id`.

## Reading the model correctly

- These are PostgreSQL schema relationships, not a claim that all services are exposed to the public web.
- The Go/PostgreSQL runtime is a local-development and CI foundation. GitHub Pages serves only the static storefront; the admin/seller services and demo-commerce write paths are internal.
- Browser-local favorites, comparison, saved searches, cart state and recently viewed history are not database tables.
- No real payment processor, inventory reservation, shipment, seller payout or money refund is connected.
- Migrations are immutable after application. Add future schema changes as forward-only migrations and follow [DATABASE_MIGRATION_POLICY.md](DATABASE_MIGRATION_POLICY.md).

## Source files

- [001–015 SQL migrations](../backend/internal/database/migrations/)
- [Backend architecture](ARCHITECTURE.md)
- [Migration and recovery policy](DATABASE_MIGRATION_POLICY.md)
