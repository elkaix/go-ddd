# Go implementation decisions

## Package boundaries

Follow the target repository's module and package conventions. If a new layout is needed, group by capability, expose a small API, and wire concrete adapters in the composition root. `internal/ordering`, `internal/payments`, and `internal/catalog` are possible capability packages, not a prescribed directory tree.

Keep domain behavior free of HTTP handlers, SQL rows, ORM lifecycle hooks, and vendor SDK types. Prefer small consumer-owned interfaces for actual dependencies; introduce an interface when a caller needs a boundary, not automatically for every struct. Translate API requests and persistence records explicitly. Read-only projections can query storage directly through an authorized query path without rebuilding write-side aggregates.

Use [`CodeReviewComments`](https://go.dev/wiki/CodeReviewComments#interfaces) for interface, context, and error conventions. Check the project's declared Go version before using language or library features.

## Value objects and aggregate mutation

Use unexported fields where callers must preserve invariants, constructors that validate input, and intent-revealing behavior. An exported Go type can still be constructed as its zero value: define whether that value is useful or invalid, then make boundary operations enforce that contract. A constructor is not an unbypassable gate.

Value receivers do not make slices, maps, or pointers immutable. Copy mutable inputs when accepting ownership and copy mutable outputs when exposing snapshots. Validate all changes before committing them to the receiver; an error should not leave a half-applied command. Declare single-owner semantics unless synchronization is deliberately provided.

Separate creation from rehydration. Creation may generate identity and apply current policy. Rehydration must preserve identity, timestamps, version, and historical facts while checking structural consistency. A newly introduced sales rule need not invalidate historical orders; define migration or versioned policy deliberately.

## Money, quantities, and time

Represent monetary amounts with a currency-aware fixed-point or decimal model. Define the currency scale and rounding policy; currencies need not share two decimal places. Validate cross-currency operations, signs, upper bounds, and intermediate overflow before mutating state. A zero-price item or order is a business decision, not universally invalid.

Represent discounts with an explicit bounded ratio, basis points, or a suitable decimal type. Test examples with known expected totals, fractional rounding, and extremes. Define whether rounding occurs per line or per order, and preserve the price/tax/discount snapshot used for the transaction. Avoid lossy float conversions for authoritative amounts.

Validate quantity units and conversion rules. For time ranges, check ordering and endpoint semantics. Model business hours in their IANA timezone and actual business date; account for overnight windows and daylight-saving transitions. Inject a clock or pass the relevant instant into deterministic rules rather than burying `time.Now()` throughout behavior.

## Errors and resources

Separate validation rejection, missing data, authorization denial, stale version, transient dependency failure, and unknown external outcome. Wrap causes with `%w` where callers need classification and use `errors.Is`/`errors.As` rather than matching messages. Translate internal errors into safe transport responses at the edge.

Take `context.Context` as the first parameter for I/O operations and pass it to the operation that actually blocks. Pure calculations generally need no context. Cancel derived contexts when finished. See [Go's cancellation guide](https://go.dev/doc/database/cancel-operations).

Give each connection pool, client, worker, and subscription a documented owner and shutdown path. A constructor returning a live resource must not defer its closure inside that constructor. Prefer injected, reusable clients/pools over per-request creation. Handle partial initialization cleanup and close response bodies/rows on all applicable paths.

For HTTP adapters, use context-bound requests, bounded deadlines and response sizes, encoded query parameters, explicit status classification, and safe endpoint configuration. Validate foreign data before translating it into domain values. Preserve correlation identifiers without logging tokens or sensitive payloads.

## Using the bundled example

Read the domain-only example under `assets/order-example/` when teaching constructor validation, value semantics, aggregate transitions, or defensive copying. Its limits are intentional: it is not a pricing engine, reservation system, repository, or payment flow. Verify its assumptions against the target use case before borrowing its design. Run its module tests separately from any parent Go module.
