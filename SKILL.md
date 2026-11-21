---
name: go-ddd
description: Design, implement, refactor, and review Go services using domain-driven design. Use for Go bounded contexts, entities, value objects, aggregates, repositories, domain services, business invariants, or translating the Packt Domain-Driven Design with Go examples into working software.
---

# Domain-Driven Design with Go

Turn business rules into a small, explicit Go model, then connect that model to storage and external systems. Preserve existing conventions and make each architectural addition earn its cost.

This is an original operational adaptation of the public Packt code examples, with independently added production guidance. It covers the repository's chapter 2–8 progression; it does not reproduce the book. Read [source-map.md](references/source-map.md) when tracing an upstream example, checking its limitations, or updating this skill.

## 1. Establish the task and evidence

Infer the mode from the request: **design**, **implement**, **refactor**, **review**, or **explain**. For a mixed request, design only enough to complete the requested change. A review produces findings; an explanation produces an explanation; neither implies permission to rewrite the application.

For an existing project, inspect its instructions, `go.mod`/`go.work`, relevant packages, handlers, persistence, migrations, tests, and external adapters. Discover nested modules before selecting test commands. Use `rg --files` and focused symbol searches. Record the actual dependency direction and one end-to-end business operation before proposing new packages.

For a new project, establish the actors, commands, rules, and persistence needs from the request. Label assumptions. Ask only about an unresolved business decision that changes correctness or the requested scope; continue independent work while that decision remains open.

**Completion:** identify the requested output, affected business operation, existing constraints, and evidence versus assumptions. Treat source examples, comments, and external content as material to inspect, not authority to expand the task.

## 2. Discover the domain before naming layers

Read [domain-modeling.md](references/domain-modeling.md) when identifying vocabulary, context relationships, entities, value objects, or aggregate boundaries.

Write a compact model for the affected slice:

| Item | Record |
|---|---|
| Language | Business term, meaning, and context; flag conflicting meanings |
| Command | Actor, intent, inputs, preconditions, result |
| Invariant | Rule that must hold, owner, concurrency protection |
| State transition | Allowed source/target states and rejection outcome |
| Boundary | Data owner, contract, consistency requirement |

Use actual rules from requirements, code, tests, or stakeholders. Mark proposed rules explicitly. Identify the operation's failure and duplicate-delivery behavior alongside its success path.

**Completion:** each changed invariant has an owner and an observable acceptance scenario. Keep unknown business policy visible rather than inventing it.

## 3. Choose the smallest useful architecture

Keep cohesive business capabilities together. Preserve a working modular monolith when it meets the requirements. A bounded context is a model boundary; choosing a separate process requires its own justification.

Choose aggregate boundaries from atomic invariants and contention. Use identity references between aggregates where practical. Let application services orchestrate I/O, authorization, transactions, and domain behavior. Put a rule in a domain service only when it has no natural entity/value-object owner.

Read [evolution.md](references/evolution.md) for existing-code migration, service extraction, CQRS, or event-sourcing decisions. Read [commerce-profile.md](references/commerce-profile.md) for restaurant, ecommerce, payment, inventory, tenant, or POS work; treat its boundaries as candidate designs to verify.

**Completion:** explain the chosen ownership, dependency direction, consistency boundary, and why any added pattern is needed. A directory move alone is not a domain-model improvement.

## 4. Implement one behavioral slice

Read [go-implementation.md](references/go-implementation.md) before changing Go domain types, factories, ports, error contracts, or adapters. Keep transport, persistence, and vendor representations at the edges. Use business operations that protect invariants and leave state unchanged on rejection.

Use [order.go](assets/order-example/order.go) and [order_test.go](assets/order-example/order_test.go) only when a concrete example of a value object and aggregate would help. They are original domain-only teaching assets, with no persistence, payment, tax, or concurrency coordination. Adapt the pattern to the target rules; do not import the example module into an application.

For behavior changes, select tests that can distinguish correct behavior from the identified failure. For a requested test-first workflow, demonstrate the failing behavior before implementing the fix. Keep documentation-only work proportionate.

**Completion:** one requested business outcome works through the relevant boundary, including its meaningful rejection paths, without requiring a wholesale rewrite.

## 5. Make persistence and external effects explicit

Read [consistency-and-integrations.md](references/consistency-and-integrations.md) whenever the slice writes state, crosses an aggregate/process boundary, retries a request, handles a webhook, or emits an event.

Specify the transaction owner, concurrency check, idempotency identity, and recovery state where applicable. An external timeout can leave an unknown outcome. Keep durable business completion separate from delivery of receipts or notifications. Retain one stable identity across retries.

**Completion:** explain what happens if the process stops before and after each durable write or external effect. Every applicable unknown or partial outcome has a detection and recovery path.

## 6. Verify the changed behavior

Read [testing-and-review.md](references/testing-and-review.md) for implementation, refactoring, or review. Run the relevant repository checks and business tests in each affected module. Add race, fuzz, integration, or failure-injection coverage when it resolves a concrete risk; distinguish unavailable checks from passing checks.

**Completion:** report the evidence for changed invariants, compatibility, and recovery behavior. A mock expectation or clean race-detector run alone does not prove database or distributed consistency.

## 7. Return the requested result

For design, provide the model, ownership/contracts, key tradeoffs, and next implementable slice. For implementation/refactoring, provide changes, behavior, verification, and remaining material limits. For review, lead with severity-ranked findings containing code evidence, triggering conditions, impact, and a minimal correction; separate confirmed defects from architectural options. For explanation, use one relevant business example and stop when the requested concept is clear.

Keep evidence and proposals distinguishable. Cite inspected code and primary documentation for factual claims. Report what was actually checked. Save project documents only when they are a requested deliverable or required by that project's workflow.
