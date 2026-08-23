# task.md — What Gets Built, In What Order

*Living document. Update status as work happens. Rules: one task in progress at a time,
a task isn't "done" until its test box is checked, and every phase ends with PLAN/goal docs updated.*
*Last updated: 2026-08-23.*

---

## Legend

- `[ ]` not started · `[~]` in progress · `[x]` done · `[-]` skipped/wontdo
- **DW** = Done When (the checkable exit condition)

## Current Status Snapshot

| Area | State |
|---|---|
| Go module (`go.mod`, module `nesh`) | ✅ exists |
| Directory structure (`cmd/`, `internal/`, `benchmarks/`, `testdata/`) | ✅ exists |
| `internal/token` (token types) | ✅ done — compiles, keyword set complete |
| `internal/lexer` | ✅ done — 98% coverage, 11 tests green |
| `internal/parser` + `internal/ast` | ✅ done — 87% coverage, precedence/errors tested |
| `internal/runtime` (evaluator) | ✅ done — 80% coverage, REPL-persistent globals |
| CLI (`cmd/nesh`) + REPL | ✅ done — script/-c/help verified, exit codes 0/1/2 correct |
| Startup benchmark | ✅ done — nesh ~10ms vs bash ~9ms warm (parity, ≤2x target met) |

**Phase 1 complete except close-out: T1.10 commit + tag `v0.1.0` pending owner approval.**
*Verified 2026-08-23: `go vet` clean, all tests pass fresh, hello.nsh output correct, error paths return proper exit codes.*

---

## Phase 1 — Foundation · target Sep 7, 2026

**DW: `nesh script.nsh` correctly runs print / let / arithmetic scripts.**

- [x] T1.1 Init Go module + project structure
  - DW: `go build ./...` succeeds on empty packages
- [x] T1.2 Token package: keywords, identifiers, strings, numbers, operators — with Line/Column from day one (Phase 3 errors depend on it)
  - DW: token types compile; keyword lookup covers let/print/if/then/else/end/fn/return/for/in/while
- [x] T1.3 Lexer: final API `New(input)`, `NextToken()` — cleaned to MODULES.md contract
  - DW: handles strings w/ escapes, ints, floats, `#` comments, significant newlines, `== != <= >= < > = + - * / ( ) ,`
- [x] T1.4 Lexer test suite — 11 tests green (tests caught + fixed a first-char-swallowing bug in `New`)
  - DW: table-driven tests covering every token type, positions, unterminated string, `go test ./internal/lexer` green
- [x] T1.5 AST node definitions (Script, Let, Print, literals, prefix/infix expressions) — all carry Pos
  - DW: compiles; nodes carry position info
- [x] T1.6 Parser: recursive descent + precedence climbing; `parser.Parse(src) (*ast.Script, *Error)`; errors carry line:col
  - DW: parses testdata/*.nsh; 7 test funcs cover AST shape, precedence, unary, print args, error positions
- [x] T1.7 Evaluator: `runtime.New(out)` — pure, Output injected, globals persist across Run (REPL); int/float arithmetic, string concat (strings only, no coercion), div-by-zero + undefined-var errors with position
  - DW: `nesh testdata/hello.nsh` prints expected output; runtime tests green
- [x] T1.8 CLI entry `cmd/nesh`: `nesh script.nsh`, `nesh -c "..."`, `-h/help`, exit codes (0 ok, 1 eval/parse, 2 usage/IO)
  - DW: all forms verified from built binary
- [x] T1.9 Minimal REPL: banner, `nesh>` prompt, globals persist, errors don't kill session, `exit` quits
  - DW: verified via piped session (history/editing = Phase 3 per plan)
- [x] T1.10 Phase 1 close-out: startup benchmark done (nesh ~10ms vs bash ~9ms warm — parity, ≤2x target met); REPL prompt hidden for piped stdin; committed + tagged `v0.1.0` (2026-08-23)

**Phase 1 ✅ COMPLETE — v0.1.0 tagged**

---

## Phase 2 — Core Language · target Sep 21, 2026

**DW: scripts with if/else, fn definitions/calls (incl. simple recursion), and PATH commands run correctly.**

- [x] T2.1 Comparison + boolean operators (`== != < > <= >= and or not`), `true`/`false` literals; truthiness rules in TECHNICAL_SPEC.md 4.4
- [ ] T2.2 `if cond then ... else ... end` execution (+ `elif`? decide and document)
- [ ] T2.3 `fn name(params) ... end` with `return`; lexical scoping decision (dynamic vs lexical — document!)
- [ ] T2.4 Function calls as expressions; recursion works
- [ ] T2.5 `for x in list` / `while cond` loops
- [ ] T2.6 Unknown command → exec from PATH, args, exit-code capture into `$?`-successor (design the idiom, e.g. `code = run git status`)
- [ ] T2.7 Unit + integration tests (>85% control-flow/fn coverage), real .nsh integration scripts in testdata
- [ ] T2.8 Phase close-out: loop/variable benchmarks vs Bash, docs updated, tag `v0.2.0`

---

## Phase 3 — AI & Polish · target Oct 5, 2026

**DW: REPL is daily-usable, errors point at line:col, agents can consume `nesh --json` output.**

- [ ] T3.1 Error system: parse + runtime errors with line/col, "expected X, got Y" messages
- [ ] T3.2 REPL: history, line editing, multi-line blocks
- [ ] T3.3 `nesh --json script.nsh`: emits JSON AST + structured execution events (the agent-API beachhead)
- [ ] T3.4 Stdlib: string ops (len, upper, lower, split, join, contains), math (abs, floor, round, min, max), file ops (read, write, exists)
- [ ] T3.5 Modules: `import "utils.nsh"` — simple, file-based
- [ ] T3.6 Memory/GC benchmark pass, docs, tag `v0.3.0`

---

## Phase 4 — Ecosystem · target Oct 19, 2026

**DW: installable, cross-platform, beats Bash in ≥70% of benchmark categories, package manager works.**

- [ ] T4.1 Benchmark suite final: startup, loops, vars, pipelines, control flow, fns, strings, math, file I/O vs bash/dash/zsh (hyperfine)
- [ ] T4.2 Profile hot paths (pprof); optimize (sync.Pool, interning, buffered I/O); document any Rust/C interop decision — only with profiling proof
- [ ] T4.3 `nesh pkg install/run` — minimal registry = git URLs
- [ ] T4.4 Cross-platform builds (macOS, Linux, Windows) in CI (GitHub Actions) + regression benchmarks
- [ ] T4.5 Docs: user guide, syntax reference, examples that all run
- [ ] T4.6 Real-world dogfooding: write 5 genuine automation scripts in Nesh, fix what hurts
- [ ] T4.7 Launch polish, tag `v1.0.0`

---

## Parking Lot (explicitly NOT now)

- Structured data / JSON literals, `fetch` — after Phase 3 core is stable
- `parallel:` blocks, `retry n:`, `task/depends_on` graphs — Phase 5+, needs concurrency design doc first
- Secrets management, remote execution — post-1.0
- Windows-first anything — we are Unix-first, Windows is a build target only

## Decision Log

| Date | Decision |
|---|---|
| 2026-08-23 | **Go over Rust** — velocity wins; OS calls dominate shell perf; Rust reserved for profiled hot paths via cgo |
| 2026-08-23 | Newlines are significant (statement terminators) — ambiguity is the enemy of agent generation |
| 2026-08-23 | Errors carry line:col from Phase 1, not retrofitted in Phase 3 |
| 2026-08-23 | **Modular architecture locked** — one-way deps, OS only via `internal/shell` interfaces, builtins as registry plugins. Contracts in MODULES.md |
