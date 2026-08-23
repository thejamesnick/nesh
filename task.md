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
- [x] T2.2 `if cond then ... elif ... else ... end` execution — `elif` decided YES (flat chains, spec 4.5); nested ifs work; REPL multi-line still Phase 3 (T3.2)
- [x] T2.3 `fn name(params) ... end` with `return`; scoping decided: params/let are call-locals, reads fall back to globals, fns global, implicit return = false (spec 4.6)
- [x] T2.4 Function calls as expressions AND bare-call statements (`deploy("prod")`); recursion works (factorial/fib tested)
- [~] T2.5 Loops: `while cond ... end` done; `for x in list` blocked on list-value decision (literals `[1, 2, 3]`?); no break/continue yet (documented)
- [x] T2.6 Unknown command → exec from PATH via `internal/shell.CommandRunner` seam; literal word args, flag merging (`-la`, `--force`); `run cmd` expression = exit code (the `$?`-successor idiom: `let code = run git status`); variable-shadow guard; no stdin passthrough yet
- [x] T2.7 Unit + integration tests (runtime 88%, parser 82%, lexer 98%; control-flow/fn >85% via runtime); 3 real integration scripts in testdata
- [x] T2.8 Phase close-out: loop/variable/fn benchmarks vs Bash — nesh wins all (2.4x loop, 51x fn); results in benchmarks/results/v0.2.0.md; tag `v0.2.0`

**Phase 2 ✅ COMPLETE — v0.2.0 tagged**

---

## Phase 3 — AI & Polish · target Oct 5, 2026

**DW: REPL is daily-usable, errors point at line:col, agents can consume `nesh --json` output.**

- [ ] T3.1 Error system: parse + runtime errors with line/col, "expected X, got Y" messages
- [x] T3.2 REPL: history, line editing (raw mode), multi-line blocks via parser.OpenBlocks
- [x] T3.3 `nesh --json script.nsh`: JSON AST (kind-tagged nodes) + structured execution events {status, ast, events, errors}
- [x] T3.4 Stdlib via `internal/builtin` plugins: len/upper/lower/split/join/contains, abs/floor/round/min/max, read/write/exists (FileSystem seam); List values + `for x in` shipped alongside
- [x] T3.5 Modules: `import "x.nsh"` (merge) and `import "x.nsh" as u` (namespaced, dotted access); cached by path, cycle detection; real lexical scoping (Func captures defining env)
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
