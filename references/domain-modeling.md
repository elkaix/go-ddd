# Domain modeling decisions

## Discover language and boundaries

Start from a business episode: who acts, what they intend, what changes, what must remain true, and who handles exceptions. Translate ambiguous words into a glossary scoped to a context. A customer in ordering may be a delivery contact while a customer in billing owns an invoice account; sharing the word does not require sharing the object.

Classify subdomains as core, supporting, or generic when that distinction helps prioritize design effort. Invest rich behavior where rules and change justify it. Simple configuration or lookup data can remain straightforward CRUD.

Propose bounded contexts where language, rules, ownership, or change cadence diverge. Verify them against use cases and team responsibilities. Avoid deriving them mechanically from database tables, HTTP resources, or a list of nouns. Record each relationship:

| Relationship | Decision to record |
|---|---|
| Upstream/downstream | Which party owns the concept and changes the contract? |
| Open host / published language | What stable API or event vocabulary is shared? |
| Anti-corruption layer | How does the consumer translate and reject foreign concepts? |
| Conformist | Is accepting the upstream model an intentional cost tradeoff? |
| Shared kernel | Which small model is jointly owned, and how are changes coordinated? |

Source basis: upstream chapter 2 and [Fowler's bounded-context explanation](https://martinfowler.com/bliki/BoundedContext.html). A context map records semantic relationships, not just network arrows.

## Choose tactical building blocks

| Building block | Use it for | Proof to seek |
|---|---|---|
| Entity | Something whose identity persists while attributes change | Two states can describe the same business identity |
| Value object | A validated value defined by its components | Equality depends on value; updates produce another value |
| Aggregate root | A consistency boundary controlling a cohesive set of state | Every mutation protects the stated invariants |
| Factory | Creation whose rules require coordination or multiple validated values | A valid initial object or an explicit rejection |
| Repository | Loading/saving an aggregate or fulfilling a focused persistence need | Business-facing contract with defined missing/conflict behavior |
| Domain service | A business policy that spans values/entities without a natural owner | Domain language and rules, independent of transport concerns |
| Application service | Executing a use case across domain behavior and ports | Clear authorization, transaction, and side-effect ownership |

Treat these as choices rather than a mandatory class inventory. An ordinary constructor can be the factory. A small Go package may contain several roles without separate layers for each.

## Define aggregates from invariants

List each invariant as a predicate with its atomic scope. For example, an order's line total and currency may be local; reserving the last unit of stock must coordinate concurrent buyers at the inventory owner. A local object check alone cannot prevent two processes from overselling.

Keep collections bounded or use a design that avoids loading an ever-growing history. Reference separate aggregates by identity where possible. If a rule seems to demand a huge aggregate, examine whether the rule truly requires immediate consistency or whether a reservation/process policy is more suitable.

Prefer one aggregate mutation per transaction as a starting point, not an absolute law. If several aggregates in one database must change atomically, record the real invariant and contention cost. If they cross autonomous stores, choose an explicit consistency protocol; calling multiple repositories does not create a transaction.

## Specify behavior before setters

Describe commands with preconditions, allowed transitions, result, and unchanged state on rejection. Keep entity identity stable across retries and rehydration. Reserve setters for data with no protected invariant; use intent-revealing methods for behavior such as reserving, accepting, canceling, or applying a validated policy.

Distinguish domain events (facts about the model), integration events (versioned external contracts), and commands (requests that may fail). Use past-tense event names and define the state that makes the event true. Raising an in-memory event does not guarantee durable delivery.

Finish with one example and one counterexample for each changed rule. This supplies acceptance tests and reveals missing policy sooner than a large class diagram.
