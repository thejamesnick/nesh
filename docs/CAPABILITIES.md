# Capabilities — What Nesh Can Do Right Now

*Snapshot of v0.4.0 (Phase 4 complete). Verified against the built binary on 2026-08-24.*

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

| Feature | Example | Notes |
|---|---|---|
| Run any PATH command directly | `git status`, `ls -la`, `docker compose up -d` | Unknown words resolve from PATH |
| Capture exit codes | `let code = run git status` | No `$?` needed |
| Capture command output | `let branch = capture git branch --show-current` | Final stage's stdout as a string (trailing newlines stripped); stderr still prints; exit code discarded — pair with `run` to check it |
| Literal args + merged flags | `-la`, `--force` | Pass through cleanly, no word splitting |
| Redirect stdout (write) | `git log > out.txt` | Absolute paths need quotes: `> "/var/log/app.log"` |
| Redirect stdout (append) | `run build >> build.log` | |
| Feed a file to stdin | `wc -l < access.log` | |
| Stdin passthrough | — | Commands read the script's own stdin when no `<` redirect is given |
| Pipelines | `cat log \| grep error \| wc -l` | stdout→stdin between stages, concurrent like a shell; exit status = last stage |
| Pipeline capture | `let n = run git log \| grep fix \| wc -l` | `run` evaluates to the last stage's exit code |
| Error handling | `try ... on failure ... end` | `fail ["msg"]` raises; handler reads the `failure` variable; bare `try` swallows; failures cross function boundaries |
| Exit codes | `exit [code]` | Sets the process exit code; not catchable by try |
| Path words | `./script.sh`, `/usr/bin/git`, `go build ./...` | Words may start with `.` and `/`; `-7` stays unary minus |

## Error Model

- **Failures** (catchable): raised by the explicit `fail` statement. Catch them with `try / on failure`.
- **Errors** (not catchable): bugs like division by zero or undefined variables. They always abort with line:col — loud by design.
- Command exit codes never abort scripts. Capture with `run` and decide:

```nesh
let code = run tests
if code != 0 then
  fail "tests broke"
end

try
  deploy("prod")
on failure
  print "rolled back:" failure
end
```

## Standard Library

| Category | Builtins |
|---|---|
| Strings | `len`, `upper`, `lower`, `split`, `join`, `contains` |
| Math | `abs`, `floor`, `round`, `min`, `max` |
| Files | `read`, `write`, `exists` (via FileSystem seam) |
| Commands | `printf` (common %s/%d/%f + \n\t), `echo [-n]` — native builtins, no process spawn |

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
print "negated:" not up
```

Actual output from the built binary:

```text
host web-01 healthy, load 0.75
status: down
negated: false
```

More runnable examples live in `testdata/`: `algorithms.nsh`, `pipeline.nsh`, `stdlib.nsh`, `importer.nsh`.

## Performance

Fastest in 4 of 7 categories vs bash/dash/zsh (see `benchmarks/results/v0.4.0.md`):

- **Wins** — loop, fn calls, control flow, redirect. The fn-call gap is structural: bash re-parses function bodies on every call; nesh compiles once (~50x faster)
- **Vars/strings** — dash wins by a hair; startup-bound (~15ms vs ~20ms boot)
- **Pipeline** — bash wins; adds ~8ms of parse/exec on top of two unavoidable external spawns
- **Startup** — ~parity with bash (~10ms warm); profiling is the Phase 5 target

---

## Not Yet (Parking Lot)

These are explicitly deferred — see `task.md` and `LIMITATIONS.md`:

- Package manager (`nesh pkg`) — Phase 5
- Maps/dicts, JSON literals, `fetch` — parking lot
- Structured pipelines (`| filter user.active`) — parking lot
- `parallel:` blocks, `retry n:`, task dependency graphs — Phase 6+
- Globbing (`*.txt`), `cd` builtin, string interpolation — see LIMITATIONS.md

## Next Phase

**Phase 5 — Share With The World**: `nesh pkg`, cross-platform CI builds,
user guide + syntax reference, then tag `v1.0.0`.
