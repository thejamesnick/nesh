# Limitations — What Nesh Can't Do Yet

*Companion to [CAPABILITIES.md](CAPABILITIES.md). Snapshot of v0.3.0, 2026-08-24.
Each item lists why it matters and where it's tracked. Nothing here is accidental —
see `task.md` (roadmap) and the Parking Lot for scheduling.*

---

## Shell Gaps

| Missing | Why it hurts | Status |
|---|---|---|
| **Pipelines (`\|`)** | Can't do `cat log \| grep error` — the Unix composability habit | Phase 5+ design; "pipelines stay sacred" per goal.md |
| **Stdin passthrough** | System commands can't read stdin from the script or terminal | Tracked in T2.6 notes |
| **Redirection** (`>`, `>>`, `<`) | No writing command output to files via shell syntax | Must use `write` builtin instead |
| **Background jobs / signals** | No `&`, no Ctrl-C handling of children, no `wait` | Not started |
| **Globbing** (`*.txt`) | Wildcards pass through literally to commands | Not started |

## Language Gaps

| Missing | Why it hurts | Status |
|---|---|---|
| **`break` / `continue`** | Loop control requires flag-variable workarounds | Documented in T2.5 |
| **String interpolation** | `"host: " + host` only — no `"...{host}..."` sugar | Not decided |
| **Maps / dicts** | Only lists exist; keyed data needs parallel lists | Parking lot (structured data) |
| **JSON literals & parsing** | Can't parse or emit JSON natively; agents/APIs need this badly | Parking lot, after core stable |
| **`fetch` / networking** | No HTTP in stdlib — must shell out to `curl` | Parking lot |
| **Error handling** (`try / on failure`) | Runtime errors abort the script; can't catch and recover | Parking lot; needs design doc |
| **`retry n:` / timeouts** | Unreliable ops need manual loops | Parking lot |
| **Parallel execution** (`parallel:` blocks) | Everything runs sequentially | Parking lot; needs concurrency design first |
| **Task dependency graphs** (`task / depends_on`) | Workflow orchestration not expressible | Parking lot |
| **User-defined errors / nil value** | Functions return `false` implicitly; no null semantics yet | Undecided |
| **Negative number literals edge cases** | Unary minus works but mixed expressions have precedence quirks | See parser tests |

## Tooling Gaps

| Missing | Why it hurts | Status |
|---|---|---|
| **Package manager** (`nesh pkg install/run`) | No way to share/install scripts from git URLs | **T4.3 — next up** |
| **Cross-platform builds/CI** | macOS/Linux only today, no Windows build target verified | **T4.4 — Phase 4** |
| **Tab completion** in REPL | Commands/paths/fns not completed | Not scheduled |
| **Syntax highlighting / editor support** | No vim/VS Code grammar for `.nsh` | Not scheduled |
| **Login-shell readiness** | Can't reliably replace bash as default shell yet (env, job control, rc files) | Post-v1.0 consideration |

## Ecosystem Gaps

| Missing | Why it hurts | Status |
|---|---|---|
| **User guide + syntax reference** | Docs are internal specs; nothing written for end users | **T4.5 — Phase 4** |
| **Real-world script validation** | Language is benchmark-proven but not dogfooded on genuine automation | **T4.6 — Phase 4** |
| **Benchmark coverage vs dash/zsh** | Current wins are vs bash only | **T4.1 — Phase 4** |

---

## Deliberate Non-Goals

These aren't limitations — they're design decisions (see goal.md):

- Not a Bash clone — we won't chase Bash compatibility
- No cryptic symbols (`$?`, `$VAR`, `${...}`) — ever
- Windows-first anything — Unix-first, Windows is a build target only
- Secrets management, remote execution — post-1.0
