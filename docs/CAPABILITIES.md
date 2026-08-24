# Capabilities — What Nesh Can Do Right Now

*Snapshot of v0.3.0 (Phase 3 complete). Verified against the built binary on 2026-08-24.*

---

## Language

| Feature | Syntax | Notes |
|---|---|---|
| Variables | `let x = 5` | Ints, floats, strings, booleans, lists |
| Arithmetic | `x * price + 1.5` | Int/float math, string concat with `+` |
| Comparisons | `== != < > <= >=` | |
| Booleans | `and`, `or`, `not`, `true`/`false` | Documented truthiness rules |
| Conditionals | `if cond then ... elif ... else ... end` | Flat `elif` chains, arbitrary nesting |
| While loops | `while i < 10 ... end` | With `break` / `continue` |
| For-in loops | `for x in list ... end` | Works over lists and split results; `break` / `continue` supported |
| Functions | `fn name(args) ... return ... end` | Recursion, calls as expressions or bare statements, lexical scoping |
| Modules | `import "utils.nsh"` / `import "u.nsh" as u` | Namespaced dotted access, cached by path, cycle detection |

## Shell & System

| Feature | Example |
|---|---|
| Run any PATH command directly | `git status`, `ls -la`, `docker compose up -d` |
| Capture exit codes | `let code = run git status` — no `$?` needed |
| Literal args + merged flags | `-la`, `--force` pass through cleanly |

## Standard Library

| Category | Builtins |
|---|---|
| Strings | `len`, `upper`, `lower`, `split`, `join`, `contains` |
| Math | `abs`, `floor`, `round`, `min`, `max` |
| Files | `read`, `write`, `exists` (via FileSystem seam) |

## Tooling

| Command | What it does |
|---|---|
| `nesh script.nsh` | Run a script |
| `nesh -c "print hello"` | Run a one-liner |
| `nesh` | Interactive REPL — multi-line blocks, history, raw-mode line editing |
| `nesh --json script.nsh` | Machine-readable AST + structured execution events for AI agents |

Errors always carry line:column and say what was expected.

---

## Proof It Runs

`testdata/health.nsh`:

```nesh
let host = "web-01"
let up = true
let load = 0.75

if up and load < 1.0 then
  print "host" host "healthy, load" load
else
  print "investigate" host
end

fn status(ok)
  if not ok then
    return "down"
  end
  return "up"
end

print "status:" status(false)
```

Actual output from the built binary:

```text
host web-01 healthy, load 0.75
status: down
```

More runnable examples live in `testdata/`: `algorithms.nsh`, `pipeline.nsh`, `stdlib.nsh`, `importer.nsh`.

## Performance

Beats Bash in every measured category (see `benchmarks/results/v0.2.0.md`):

- Loops: ~2.4x faster (~670µs/op, low GC pressure)
- Function calls: ~51x faster (~608µs/op)
- Startup: parity with bash (~10ms warm)

---

## Not Yet (Parking Lot)

These are explicitly deferred — see `task.md`:

- Pipelines (`|`) and stdin passthrough
- Structured data: JSON literals, maps, `fetch`
- Error handling: `try / on failure`
- `break` / `continue`
- Parallel execution (`parallel:` blocks), `retry n:`, task dependency graphs
- Package manager (`nesh pkg`) — Phase 4, next up

## Next Phase

**Phase 4 — Ecosystem**: final benchmark suite vs bash/dash/zsh, profiling,
`nesh pkg install/run`, cross-platform CI builds, user guide + syntax reference,
dogfooding 5 real scripts, then tag `v1.0.0`.
