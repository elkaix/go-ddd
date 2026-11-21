# Current Go baseline and engineering practices

## Verify release freshness

**Verified 2026-11-21:** latest stable patch **Go 1.27.4**, released **2026-11-10**. The [Go 1.27.0 release anchor](https://go.dev/doc/devel/release#go1.27.0) identifies the release family, initially released 2026-08-19. Patch releases include security fixes; prefer the latest supported patch after checking compatibility.

Use the [official release history](https://go.dev/doc/devel/release), [stable download feed](https://go.dev/dl/?mode=json), and [Go 1.27 release notes](https://go.dev/doc/go1.27) as the authorities. This dated snapshot is not a permanent definition of latest. Before recommending an upgrade or describing current behavior, recheck these sources, select a stable release rather than a beta/RC, record the date, and inspect intervening patch notes. If sources cannot be reached, state the last verified baseline and the freshness limitation.

The [release policy](https://go.dev/doc/devel/release#policy) supports a Go release until two newer major releases exist. The supported families at this snapshot are 1.27 and 1.26; verify support again during future work.

## Separate language, compiler, and deployment choices

Inspect every relevant `go.mod`, `go.work`, CI setup, container base image, and deployment build. Record `go version` and targeted `go env` values such as `GOVERSION`, `GOTOOLCHAIN`, `GOWORK`, and `GOEXPERIMENT`; avoid dumping unrelated environment variables.

The [`go` directive](https://go.dev/doc/modules/gomod-ref#go) declares the minimum required Go version and controls language-version semantics. The [`toolchain` directive](https://go.dev/doc/modules/gomod-ref#toolchain) suggests a compiler for the main module/workspace; it is not an exact lock and does not require downstream consumers to use that suggestion. [`GOTOOLCHAIN`](https://go.dev/doc/toolchain) and workspace selection can change which compiler actually runs.

For a new application, choose the current stable patch unless requirements dictate another supported family. For an existing project or library, preserve its compatibility promise until the requested change includes a migration. Record a deliberate minimum-version increase, review dependency requirements, and test supported platforms. Keep production CI/builds pinned to the selected patch; a separate moving-stable check can serve as compatibility surveillance.

The bundled example intentionally targets `go 1.27.0`, suggests `toolchain go1.27.4`, and is tested in the public repository with Go 1.27.4. Recheck and update these together when maintaining this skill. They do not authorize unrelated applications to raise their language floor.

## Apply Go 1.27 features where they help

The [release notes](https://go.dev/doc/go1.27) document these changes. Confirm exact APIs in version-matched package documentation before writing code.

| Change | Application decision |
|---|---|
| Generic methods | Require Go 1.27 language support. Interface methods still cannot declare type parameters or be implemented by generic methods. Keep domain ports small. |
| `encoding/json/v2` and `encoding/json/jsontext` | Evaluate explicit adoption at wire/storage boundaries. V2 rejects duplicate object names and invalid UTF-8 by default; test compatibility with clients and stored payloads. |
| Existing `encoding/json` uses the new implementation | Its intended behavior remains compatible, but error text can change. Assert error meaning and contract behavior rather than exact message text. |
| Standard `uuid` package | Evaluate against the required ID formats and persistence contracts; preserve existing IDs and database/wire representations during migration. |
| `goroutineleak` profile | Use it when diagnosing permanently blocked goroutines; detection is incomplete. Restrict profiling endpoints to authorized diagnostic access. |

Review selected `go fix` modernizations in an isolated change and inspect the diff. Treat experimental SIMD APIs as optional experiments, not default application dependencies. A newer feature earns adoption through simpler code, correctness, or measured performance.

## Idiomatic boundaries and errors

Use [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) for small consumer-owned interfaces, names, contexts, and error style. Use [Effective Go](https://go.dev/doc/effective_go) for foundational idioms alongside current release/module documentation; its own introduction warns that it does not cover the modern ecosystem comprehensively.

Preserve explicit ownership and business operations from the implementation guide. Prefer standard-library capabilities when they satisfy the contract; introduce libraries or generic abstractions for a concrete need. Keep value equality, zero-value behavior, mutable ownership, and rehydration invariants explicit.

Use `errors.Is` for sentinel matching and `errors.As` for typed extraction through wrapped chains. On Go 1.26+, [`errors.AsType[E]`](https://pkg.go.dev/errors#AsType) provides typed extraction with a boolean result; use it only when the project's minimum supports it. Preserve error causes needed by callers and expose safe domain/transport categories. Avoid blanket rewrites of working error code merely to change syntax.

## Bound concurrency and own its lifetime

Give each goroutine an owner, stop condition, cancellation path, and join point. Bound work admission and memory/connection use; pass cancellation to the blocking operations themselves. Do not detach a request merely to make a side effect finish: use the durable application workflow when work must survive the request.

On Go 1.25+, [`sync.WaitGroup.Go`](https://pkg.go.dev/sync#WaitGroup.Go) manages task counting, but the function must not panic and the method does not propagate errors or bound concurrency. For error-producing concurrent tasks, consider [`errgroup.WithContext` and `SetLimit`](https://pkg.go.dev/golang.org/x/sync/errgroup). Set the limit before launching, leave it unchanged while tasks run, and make functions honor cancellation. Admission through `Go` can block at the limit; design the producer's cancellation behavior explicitly. Always wait for completion; the derived context is canceled when `Wait` returns, including success, so use the appropriate parent/new context for subsequent work.

Follow the [Go memory model](https://go.dev/ref/mem) for shared state. Prefer clear ownership or appropriate synchronization over assumptions about goroutine scheduling. The race detector complements this reasoning; it cannot prove absence of all concurrency defects.

## Test behavior and measure performance

Use the existing testing guide for domain and integration coverage. On Go 1.25+, [`testing/synctest`](https://pkg.go.dev/testing/synctest) can provide deterministic time and synchronization for suitable in-process concurrency tests. Its bubble model does not replace real database, filesystem, or provider integration tests; ensure all started work terminates.

For libraries retaining an older consumer minimum, supplementary tests can use [version build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints), such as `//go:build go1.25`, while preserving equivalent baseline coverage. Verify the promised minimum with that actual compiler and `GOTOOLCHAIN=local`, alongside a supported compiler; automatic toolchain switching must not conceal incompatibility.

On Go 1.24+, consider [`testing.B.Loop`](https://pkg.go.dev/testing#B.Loop) for new benchmarks. Measure representative workloads and allocations, compare repeated samples, and use [profiling](https://go.dev/doc/diagnostics) to find bottlenecks before optimizing. Avoid selecting architecture or claiming performance from a single synthetic result.

Use [fuzzing](https://go.dev/doc/tutorial/fuzz) for parsers, arithmetic, serialization, and transition properties. Keep tests bounded and deterministic where possible. Add race, contract, fuzz, or load coverage in response to a concrete changed risk rather than as an automatic full-suite expansion.

## Dependencies, tools, and security

Review dependency and module changes using the [module reference](https://go.dev/ref/mod). Keep `go.mod`/`go.sum` changes attributable, check the selected module graph, and review `replace` directives. Preserve normal checksum verification and use scoped private-module configuration where needed. Pin CI tools and actions to reviewed versions; use the project's existing tool-dependency convention, including `tool` directives when appropriate for its Go version.

For modules published independently, include `GOWORK=off` verification so workspace replacements cannot mask missing dependencies or unsupported requirements. Use workspace files deliberately for repositories whose modules are developed together, rather than assuming every local workspace belongs in version control.

Use [`govulncheck`](https://go.dev/doc/security/vuln/) to assess known vulnerabilities reachable from the analyzed Go code, including relevant standard-library issues. Run it for dependency/toolchain changes and release gates using a reviewed tool version. Record the toolchain, build tags, target platform, and database freshness; investigate findings and retest the affected behavior. A clean scan is not a guarantee of security, and an unavailable database/tool is not a passing result.

The public repository pins `govulncheck` **v1.10.0**, verified on 2026-11-21. Its [command documentation](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) distinguishes output modes: the default text mode fails on findings, while JSON, SARIF, and OpenVEX output can exit successfully despite findings. Keep a real failure gate when generating structured reports. Recheck the tool version and its minimum Go requirement during future updates.

For a targeted upgrade: establish baseline tests; review release and security notes; update the chosen compiler and compatible module/tool settings; inspect `go mod tidy` changes; check formatting, vet, relevant tests/race runs, and vulnerability findings; then record compatibility and rollback limits. Add wire-contract and benchmark comparisons when the changed runtime/library behavior makes them relevant. Verify actual output before reporting success.
