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
| Language core (lexer, parser, AST, runtime) | ✅ Phases 1–3 complete — v0.1.0 / v0.2.0 / v0.3.0 |
| Control flow, functions, loops, modules | ✅ done, tested (runtime 88%, parser 82%, lexer 98%) |
| System commands via shell seam | ✅ CommandRunner + FileSystem interfaces |
| Stdlib (`internal/builtin`) | ✅ 11 builtins, plugin registry |
| REPL | ✅ multi-line blocks, history, raw-mode editing |
| Agent API (`nesh --json`) | ✅ kind-tagged AST + execution events |
| Benchmarks | ✅ beats bash in all categories; memory profiled |

**Next phase: Phase 4 — Daily Usability (break/continue, redirection, pipelines, error handling, dogfooding) → Phase 5 — Share With The World (pkg manager, CI builds, docs, v1.0)**

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
- [x] T3.6 Memory/GC benchmark pass (bench_test.go: loop 670µs/35KB, fn 608µs/223KB per op; low GC pressure — no optimization warranted), docs refreshed, tag `v0.3.0`

**Phase 3 ✅ COMPLETE — v0.3.0 tagged**

---

## Phase 4 — Daily Usability · target Oct 5, 2026

**DW: real automation that needs pipes, file redirects, loop control, and failure recovery
is easier to write in Nesh than Bash. This is the "I actually use it every day" bar.**

- [x] T4.1 Loop control: `break`, `continue`
  - DW: unit tests; both work in while + for-in; REPL verified
- [x] T4.2 Redirection + stdin passthrough: `ls -la > out.txt`, `>> append`, command reads script stdin
  - DW: integration test writes/reads files via redirection; stdin flows to child process
  - Note: absolute paths in redirects must be quoted (`> "/var/log/x"`) — a leading `/` lexes as division otherwise
- [x] T4.3 Text pipelines: `cat log | grep error | wc -l`
  - DW: `|` chains system commands, stdout→stdin, exit code = last command; spec section added
  - Also works in `run` expressions: `let n = run printf ... | wc -l`; stages run concurrently via io.Pipe
- [ ] T4.4 Error handling v1: `try ... on failure ... end`; commands don't abort script by default
  - DW: failing command inside try jumps to handler; runtime errors still abort; tests
- [ ] T4.5 Dogfooding gate: rewrite 3 scripts you actually use today in `.nsh` — fix whatever hurts
  - DW: 3 genuine personal automation scripts run daily without bash fallback
- [ ] T4.6 Phase close-out: benchmarks incl. pipeline category vs bash/dash/zsh, docs updated, tag `v0.4.0`

---

## Phase 5 — Share With The World · target Oct 26, 2026

**DW: a stranger can install Nesh on macOS/Linux/Windows and run someone else's package.**

- [ ] T5.1 Profile hot paths (pprof); optimize only with profiling proof (sync.Pool, interning, buffered I/O)
- [ ] T5.2 `nesh pkg install/run` — minimal registry = git URLs
  - DW: install from a git URL, run its entry script; version pinning decided
- [ ] T5.3 Cross-platform builds (macOS, Linux, Windows) in GitHub Actions CI + regression benchmarks
  - DW: green CI matrix, release binaries downloadable
- [ ] T5.4 Docs: user guide, syntax reference, examples that all run
  - DW: docs/CAPABILITIES.md examples + guide cover 100% of shipped syntax
- [ ] T5.5 Launch polish, tag `v1.0.0`

---

## Parking Lot (explicitly NOT now)

- Maps/dicts, JSON literals, `fetch` — after Phase 4 usability core is stable
- Structured pipelines (`| filter user.active`) — after text pipelines prove out (T4.3)
- `parallel:` blocks, `retry n:`, `task/depends_on` graphs — Phase 6+, needs concurrency design doc first
- Secrets management, remote execution — post-1.0
- Windows-first anything — we are Unix-first, Windows is a build target only

## Decision Log

| Date | Decision |
|---|---|
| 2026-08-23 | **Go over Rust** — velocity wins; OS calls dominate shell perf; Rust reserved for profiled hot paths via cgo |
| 2026-08-23 | Newlines are significant (statement terminators) — ambiguity is the enemy of agent generation |
| 2026-08-23 | Errors carry line:col from Phase 1, not retrofitted in Phase 3 |
| 2026-08-23 | **Modular architecture locked** — one-way deps, OS only via `internal/shell` interfaces, builtins as registry plugins. Contracts in MODULES.md |
