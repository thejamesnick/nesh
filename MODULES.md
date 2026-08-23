# MODULES.md — Package Contracts and Dependency Rules

*The modularity law of Nesh. A PR that breaks a rule here gets rejected — no exceptions,
including "it was faster". Last updated: 2026-08-23.*

---

## The One-Way Dependency DAG

Dependencies may only point downward. Ever.

```
cmd/nesh                 ← wiring only, no language logic
   │
internal/runtime         ← evaluator: values, scopes, execution
   ├── internal/ast      ← node definitions, position info
   ├── internal/parser   ← tokens → AST
   ├── internal/builtin  ← stdlib plugins (registered into runtime)
   └── internal/shell    ← the ONLY package allowed to touch the OS
          │
internal/lexer           ← source → tokens
   │
internal/token           ← token types (leaf; imports nothing of ours)
```

**Forbidden:** upward imports, sideways imports between siblings, cycles. If package A needs
what package B has, either B moves down or an interface moves up.

---

## Package Contracts

### `internal/token` — the vocabulary
- Exports: `Type`, `Token{Type, Literal, Line, Column}`, `LookupIdent`.
- Imports: nothing of ours. Pure data.
- Contract: knows nothing about lexing, parsing, or execution.

### `internal/lexer` — source to tokens
- Exports: `New(input string) *Lexer`, `NextToken() token.Token`.
- Imports: token only.
- Contract: pure function of input → tokens. No I/O, no config, no globals.
  Every token carries line/column. Newlines are significant (statement terminators).

### `internal/ast` — the tree
- Exports: node types (Script, Command, Let, If, Fn, expressions…), all carrying position.
- Imports: token only.
- Contract: dumb data. No behavior beyond String() helpers for debugging.

### `internal/parser` — tokens to tree
- Exports: `Parse(tokens) (*ast.Script, error)`.
- Imports: token, ast. Drives the lexer (or accepts a token stream).
- Contract: never executes anything, never touches the OS. Errors carry position.

### `internal/runtime` — the evaluator
- Exports: `Runtime` with `Run(script)`, value types, scope.
- Imports: ast, and the *interfaces* defined in `internal/shell`.
- Contract: pure Go — no `os/exec`, no filesystem, no env, no time. Everything external
  arrives through injected interfaces, which is what makes the whole language unit-testable
  without faking a single syscall.

### `internal/shell` — the OS adapter (the seam)
- Defines the interfaces the runtime speaks to:
  - `CommandRunner` — spawn processes, capture exit codes (backs `git status`, future `run`)
  - `FileSystem` — read/write/exists/move (backs Phase 3 file builtins)
  - `Env` — environment variables
  - `Clock` — time (backs future `retry`/`timeout`; no test ever sleeps again)
- Ships the real implementations (`RealRunner`, `RealFS`, …). Tests use fakes.
- Contract: the ONLY package that may import `os`, `os/exec`, `syscall`, `net`.
  This is the entire cross-platform story — Windows support means implementing these
  interfaces for Windows, not scattering `runtime.GOOS` checks through the evaluator.

### `internal/builtin` — the stdlib, as plugins
- Exports: `Builtin` interface + `Registry`. `print`, `len`, string/math/file functions
  each live in their own file, self-registering or registered by `cmd/nesh`.
- Contract: adding a stdlib function = adding one file. Zero changes to runtime.
  Builtins get OS access only via the injected shell interfaces — never directly.

### `cmd/nesh` — the composition root
- Contract: builds the object graph — real shell implementations → runtime ← builtins ← REPL/CLI —
  then gets out of the way. If main.go contains language semantics, the code is in the wrong package.

---

## Rules of Engagement

1. **API ≤ one screen.** If a package's public surface doesn't fit on screen, split the package.
2. **Interfaces live with the consumer**, implementations with the provider — runtime defines
   what it needs; shell provides it.
3. **Every interface gets a fake** in its test package, before its second consumer exists.
4. **No star-importing around the DAG.** `runtime` may not import `lexer`. `parser` may not import `builtin`.
5. **New feature = new file (ideally new package), not an edit thread through five packages.**

## Future Seams (designed for, not built yet)

| Seam | Unlocks | Phase |
|---|---|---|
| `runtime.Value` interface | structured data, JSON values, structured pipelines | 3–5 |
| `ModuleLoader` interface | `import "utils.nsh"` | 3 |
| Event stream from runtime | `nesh --json` agent API — structured execution events | 3 |
| `PackageSource` interface | `nesh pkg` (git URLs now, registry later) | 4 |

## Known Open Items (documented, not accidental)

- Lexer currently folds `.`, `-`, `/` into identifiers (`src/main`, `user.name` lex as one token)
  — arithmetic with spaces (`a / b`) is unaffected. Final call happens in T2.6 when PATH
  commands + flags get their parser design.
