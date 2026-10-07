# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`go-core` is a Go library (module `github.com/mondegor/go-core`, Go 1.26) providing
infrastructure for building user/system/internal error handling, multi-language message
formatting and localization, and `slog`-based logging. It is a library, not an application —
there is no main binary (the only `cmd` is the `gotext-catalog-fix` codegen helper). Most
documentation and code comments are in Russian.

## Commands

Day-to-day commands are wrappers around the external [`mrcmd`](https://github.com/mondegor/mrcmd)
tool (`go-dev` plugin). Convenience `make` targets exist; use plain `go` when `mrcmd` is unavailable.

- `make test` / `mrcmd go-dev test` — run all tests
- `make test-report` / `mrcmd go-dev test-report` — tests + HTML coverage (`test-coverage-full.html`)
- `make lint` — fmt (gofumpt), goimports, gci, then golangci-lint (`.golangci.yaml`); `mrcmd go-dev lint` runs only golangci-lint
- `make generate` — run `go:generate` (gotext localization catalogs in `examples/mrlocale/dict/*`);
  requires `gotext` and `gotext-catalog-fix` in `PATH` (`go install ./cmd/gotext-catalog-fix`)
- `make plantuml` — render `.puml` diagrams under `docs/` to images

Without `mrcmd` installed:
- All tests: `go test ./...`
- Single package: `go test ./errors/runtime/`
- Single test: `go test -run TestName ./errors/user/`
- Benchmarks (several `compare_std_bench_test.go` files): `go test -bench=. ./errors/userfast/`

`golangci-lint` runs with `default: all` and a curated disable list in `.golangci.yaml` — it is
strict. Dev tools (`gofumpt`, `goimports`, `gci`, `golangci-lint`, `gotext`, ...) are installed via
`mrcmd go-dev install-tools`. `gofumpt`, `goimports`, `gci`, `golangci-lint` default to `latest`
(pin a version by uncommenting its `GO_DEV_TOOLS_INSTALL_*` in `.env`); `mockgen` and `gotext` are
disabled in go-dev by default, their versions are set in `.env.dist`.

## Architecture

### Error system (`errors/`)

The core of the library. Errors are classified by **kind** (`errors/kind`): `Internal` (developer
bugs), `System` (infra/network failures), `User` (user-correctable input). `kind.Analyze` walks
the wrap chain to determine the effective kind — a `User` error wrapping a `System`/`Internal`
cause resolves to that underlying kind.

The package is built in layers, with a flat facade re-exporting the subpackages:

- **`errors/` (root)** — the public facade. Files like `runtime_error.go`, `user_error.go`,
  `custom_error.go`, `error_wrappers.go` define type aliases (e.g. `RuntimeProtoError = runtime.ProtoError`)
  and constructors (`NewInternalProto`, `NewUserProto`, `NewUserError`, `WithCustomCode`). Prefer
  importing the root package; reach into subpackages only when extending behavior. Predefined error
  catalogs live in `internal_errors.go`, `system_errors.go`, `user_errors.go`, `event_errors.go`.
  `std.go` re-exports `New`/`Is`/`As`/`Unwrap`/`Join` from the stdlib, so this package replaces the
  standard `errors` import. `custom_wrappers.go` provides `CustomWrapper` (`NewCustomWrapper`,
  `NewDownloadFileWrapper`, `NewDownloadImageWrapper`, `NopCustomWrapper`) for user errors with a custom code.
- **`errors/runtime/`** — `ProtoError` for `Internal`/`System` errors. A *proto* is a factory:
  you build it once with text/options, then call `New`, `WithDetails`, `Wrap`, `WithError` to mint
  concrete instances carrying attrs (key/value pairs for logging), an optional `hint`, and a wrapped
  cause. `WithOnCreate` (see `options.go`) attaches per-instance data such as an error ID + stack trace.
  Supporting: `runtime/stacktrace` (caller capture), `runtime/instance` (ID generation), `runtime/hint`.
  Protos created via `NewDelayed` receive default options later: `InitDelayedOptions(handler)` (one-shot,
  called by `wire/errors.InitErrors`) applies them to already created and all subsequent delayed protos,
  never overriding options explicitly passed to `NewDelayed` (`applyDefaultOptions` copies only unset
  fields — extend it when adding a new `Option`).
- **`errors/user/`** — `ProtoError` for user errors *with* placeholder arguments.
- **`errors/userfast/`** — lighter user errors *without* arguments (root `UserError` aliases this).
- **`errors/custom/`** — wraps a `User` error with an extra "custom code" (e.g. a field name).
- **`errors/wrap/`** — `ErrorWrapper` strategy objects composed for use at architectural boundaries.
  `NewShellErrorWrapper(wrapFunc, defaultWrapper)` tries `wrapFunc` first, falling back to a default
  (typically `NewKindlessErrorWrapper`). The root `error_wrappers.go` assembles layer-specific
  wrappers (`NewInfraStorageWrapper`, `NewServiceOperationFailedWrapper`, etc.) — these encode the
  intended usage: infra-layer code wraps unrecognized errors into storage errors, service-layer code
  into operation-failed errors.
- **`errors/helper/`** — `ErrorStatusMapper` (map user error codes → HTTP statuses), property extraction.
- **`errors/handler/`** — `Handler` interface for centralized error routing (logging/tracing),
  with a `Func` adapter and `Nop()`. The root `errors/handler.go` re-exports these as `Handler`,
  `HandlerFunc`, and `NopHandler`. The library only provides this integration point — the application
  implements it and plugs in its own logging/tracing/monitoring.

The intended flow (see `errors/README.md`): each architectural layer defines its own errors and
wraps caught errors as it propagates them up; the UseCase layer is the primary interception point
that classifies the final kind and decides handling.

### Other packages

- **`mrmsg/`** — placeholder-based message formatting with configurable delimiters (e.g. `{` `}`
  or `{{` `}}`). `PlaceholderExtractor` finds placeholders; `PlaceholderReplacer.Replace(args []any)`
  substitutes positional args; `MessageFormatter` rewrites placeholders via a callback. Subpackage
  `templater`.
- **`mrlocale/`** — localization (`golang.org/x/text` + gotext catalogs): bundles, pools, localizers.
- **`mrlog/`** — `slog`-based logging (`logger.go`, `nop_logger.go`, `std.go`, timezone and
  datetime format helpers); subpackages `slog`, `level`, `color`.
- **`mrtrace/`** — `Tracer`/`ContextManager` interfaces (`NopTracer`), request/worker/correlation/task
  context propagation and ID generation (subpackages `context`, `process`).
- **`wire/`** — composition-root helpers that assemble errors + logging + tracing together
  (subpackages `errors`, `mraccess`, `mrlog`, `mrtrace`). These are default factories and usage
  examples, fully replaceable by the application. `wire/errors`: `InitErrors(ErrorConfig)` (hint with
  error ID for System/Internal, plus stack trace for Internal) and `InitErrorHandler(logger)`.
- **`mrtype/`**, **`mrmodel/`**, **`mrpath/`**, **`mrevent/`**, **`mrinfra/`**, **`mrapp/`**,
  **`mrentity/`** — supporting domain types (paging/sort/parse, file/image models, etc.).
- **`mrworkflow/`** — status flow maps (`flow_map.go`, `itemstatus`).
- **`mraccess/`** — access control (actions, rights, roles, role groups).
- **`mrstorage/`**, **`mrpostgres/`** — storage abstractions and the PostgreSQL implementation
  (connection manager, query builders, monitoring, migrations).
- **`mrprocess/`** — background processing: workers, jobs, schedulers, consumers, collectors.
- **`mrrun/`** — application runner / health probes / graceful shutdown.
- **`mrlock/`** — distributed locking (`locker.go`, `mutexlocker`, `noplocker`).
- **`mridempotency/`** — idempotency providers/responsers.
- **`util/`** — standalone helpers (e.g. `xtime`, `xmath`, `xstrings`, `ximage`, `xio`, `slices`,
  `conv`, `casttype`, `copyptr`, `crypt`, `mime`, `args`, `timezone`).

### Reference material

- `examples/` contains runnable usage examples per subsystem (`examples/errors/*`, `examples/mrlocale`,
  `examples/mrlog`, `examples/mrstorage`, `examples/mrworkflow`, `examples/shutdown`, `examples/util`)
  — the best starting point for understanding intended API usage.
- `docs/` holds C4 architecture diagrams: PlantUML sources (`docs/packages/c4/errors.puml`,
  `wire_errors.puml`, `docs/diagrams/c4/hld.puml`, shared `.iuml` parts in `docs/components/c4/`) and
  SVGs in `docs/resources/` referenced from `errors/README.md`. SVGs are rendered by `make plantuml` and
  committed together with the sources.
- `errors/README.md` (Russian) has the authoritative explanation of the error-handling philosophy.

## Conventions

- Exported symbols are documented in **English** or **Russian**; match that style when editing existing packages.
- The "Proto" pattern is pervasive: a proto error is an immutable factory; concrete errors are
  derived from it. Don't mutate protos after construction.
- Files end with a trailing newline (enforced; recent commits fixed this) and code must pass the
  strict `golangci-lint` config — run `make lint` before considering work complete.
- Project skills live in `.claude/skills/`: `go-style-guide` (conventions for writing Go code),
  `check-go-style` (report style violations), `audit-go-package` (deep correctness/robustness audit).
