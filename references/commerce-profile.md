# Optional commerce and POS profile

Use this profile for restaurant commerce, ecommerce, payment, inventory, multi-tenant, and POS requests. These are design candidates and review questions, not claims about MAD Seafood or any connected repository. Inspect the actual system before assigning ownership or proposing a migration.

## Candidate capabilities

| Capability | Possible ownership | Boundary question |
|---|---|---|
| Catalog | Product definitions, modifiers, sellability | Which source owns each field: merchant tooling or POS? |
| Pricing/promotions | Offer eligibility, rounding, price decisions | Which version applies, and is the accepted price snapshotted? |
| Ordering | Accepted line snapshot, customer intent, order lifecycle | Which changes remain legal after placement/payment? |
| Payments | Payment attempts, provider outcomes, refund accounting | How are unknown outcomes reconciled without a duplicate charge? |
| Inventory/capacity | Reservations, availability, expiry | What atomic mechanism prevents competing claims? |
| Fulfillment | Kitchen, pickup, delivery, readiness | Which actor/event can advance each transition? |
| Customer/identity | Accounts, access, preferences | What ownership and consent rules apply to each command? |
| POS integration | Mapping, sync state, external references | How do provider concepts become internal commands/facts? |
| Notifications | Delivery attempts, templates, consent enforcement | Can failure be retried independently of the purchase? |

Merge or split capabilities based on actual language, rules, ownership, and scale. This table does not require nine services or nine aggregates. A restaurant can retain all of them in one Go deployment with intentional interfaces.

## Ownership and tenant scope

Create a field-level source-of-truth map for catalog, prices, tax, stock, orders, payment status, and fulfillment. Avoid two systems both treating themselves as unconditional authority. Record reconciliation policy for drift and preserve external identifiers separately from internal identity.

Derive tenant/merchant/location scope from authenticated, authorized context and apply it to repository lookups, cache keys, idempotency keys, unique constraints, events, jobs, and provider credential selection as appropriate. Never trust an arbitrary request tenant ID as proof of access. Verify callbacks belong to the correct installation and resource before applying transitions.

Keep a POS provider such as Clover behind a small anti-corruption layer. Define ports from application needs instead of mirroring the entire vendor SDK. Verify the current provider documentation for the exact product, API, region, permissions, webhook authentication, payment/refund support, rate limits, and idempotency behavior. This profile makes no claim that any particular endpoint or guarantee is available.

## Critical business invariants

Specify authoritative prices and quantity units; validate modifiers and availability server-side. Preserve the accepted monetary snapshot with explicit currency and rounding. Model reservation expiry and order/payment state together so a late payment cannot silently consume released inventory or exceed capacity.

Distinguish payment authorization, capture, failure, uncertainty, and refund states where the provider supports them. Bound cumulative refunds against confirmed captured/refundable amounts with concurrency protection. Preserve a durable operation identity and provider reference through retries and callbacks.

Define whether identical lines merge and which attributes make them identical. Test cart/order limits before mutation. An order accepted by the server must remain identifiable even if a client disconnects or loses local storage. A receipt failure must not turn a completed charge into a fresh checkout operation.

Model cancellation versus payment success as a race with a reconciliation policy. Keep terminal states and late events explicit. Record who may issue refunds, override capacity, change credentials, or resolve stuck operations, with authorization and an audit trail appropriate to the product.

## High-value acceptance scenarios

Exercise double-click checkout; retry after a response is lost; same key with different cart data; concurrent final-stock purchase; callback before synchronous response; duplicate/out-of-order callback; successful charge followed by database failure; receipt delivery failure; cancellation during capture; refund retry after timeout; credential change during queued sync; cross-tenant order ID; and recovery after worker restart.

Choose only scenarios relevant to the requested slice. Connect each selected scenario to its invariant owner, persisted evidence, and recovery outcome. For all external effects, apply the consistency reference linked from the main skill.
