# M09.2 — Atomic vendor split checkout

`Service.CheckoutVendorCart` is an internal-only checkout path for carts containing vendor offers. It is deliberately separate from the legacy `Checkout` method; mixed carts are rejected until an explicit platform-offer mapping is designed.

## Transaction flow
1. Lock the cart and perform idempotency replay lookup before lifecycle checks.
2. Reject expired/closed carts and any legacy line mixed into the vendor cart.
3. Lock all offer and vendor rows in deterministic vendor/offer order.
4. Revalidate active vendor, active offer, variant state and cart currency.
5. Count unexpired active reservations per vendor offer and reject insufficient availability.
6. Create the parent demo order and one child order per vendor.
7. Snapshot canonical vendor offer price/title and create linked per-offer inventory reservations.
8. Calculate coupon eligibility and redemption under the coupon row lock.
9. Allocate the parent discount proportionally to child subtotals, with the final child absorbing cent-rounding remainder so child totals exactly reconcile to the parent.
10. Append the parent event, close the cart and commit atomically.

## Lifecycle propagation
- Simulated payment completion marks all pending child orders `paid_simulated`.
- Owner cancellation releases reservations and marks pending child orders `cancelled`.
- Reservation expiry marks the parent and pending child orders `expired`, expires active reservations and appends the parent event.
- These are state-only demo transitions. No payment gateway, capture, settlement, fulfillment or stock decrement is implemented.

## Guardrails and limitations
- Vendor checkout never reads or decrements legacy `products.stock`; reservations are scoped to `vendor_product_id`.
- Same-offer checkout races serialize on the offer row; active reservations are checked while that row is locked.
- One vendor child order is created per parent/vendor pair.
- Coupons redeem once at parent-order level; child discount allocations sum exactly to the parent discount.
- The current coupon per-customer cap remains based on the guest cart capability digest, not verified identity.
- Mixed legacy/vendor carts are rejected. HTTP routes, seller authorization, browser session cookies and production infrastructure remain disabled.
