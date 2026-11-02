# M09 — Vendor-aware cart and order foundation

Migration 006 adds a vendor-aware extension while preserving all legacy rows and checkout semantics.

## Cart
- `cart_items.vendor_product_id` identifies the exact vendor offer. A composite foreign key also verifies that the offer belongs to the same catalog product.
- Legacy unique indexes now apply only to rows with no vendor offer. A separate cart/offer unique index allows the same product to appear as separate offers from multiple sellers.
- Internal methods `AddVendorItem`, `SetVendorItemQuantity` and `RemoveVendorItem` lock the cart and relevant offer, require an active vendor and active offer, and snapshot canonical offer price/title/currency.
- `GetCart` returns offer/vendor identity and display name. It never exposes capability tokens.
- The old catalog `AddItem`, quantity and removal methods explicitly target legacy rows only.

## Order structure
- `demo_vendor_orders` introduces one vendor child order per parent order/vendor pair.
- Order items and inventory reservations can reference both the vendor offer and its child order. Pair constraints prevent mixing vendor and legacy ownership fields.
- Legacy order rows remain nullable and continue to use the existing single-catalog workflow.

## Checkout safety boundary
The existing `Checkout` method deliberately rejects carts containing vendor offer lines with `ErrVendorCheckoutNotReady`. It must never price or reserve those lines using the legacy product inventory. The new child-order schema is groundwork only; vendor-aware split checkout, coupon allocation, payment-state propagation and cancellation propagation remain the next milestone.
No HTTP routes, seller identity, browser cookies, real payments or production resources are enabled.
