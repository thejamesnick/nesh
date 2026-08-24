# goal.md — What We're Building and Why

*The single source of truth for direction. If a decision conflicts with this file, this file wins.
Last updated: 2026-08-24.*

---

## Mission

Build **Nesh** — a modern, human-readable shell and scripting language for the agent era.
Same job as Bash (automation, scripting, interactive shell, default-shell capable), but readable
without a manual and reliably generatable by AI agents.

> The same properties that make code readable for humans make it reliable for AI.

## Identity

| Thing | Name |
|---|---|
| Language | Nesh |
| Runtime / CLI | `nesh` |
| Script extension | `.nsh` |
| Implementation language | **Go** (locked 2026-08-23 — Rust only as cgo hot-path escape hatch, if profiling ever demands it) |
| License | MIT |

## The Three Users We Serve

1. **Developers** writing automation who are tired of Bash quoting landmines
2. **AI agents** generating and executing shell scripts — they need unambiguous, predictable grammar
3. **Reviewers** (same humans) who must trust what the agent wrote in seconds

## What Nesh Is

- A full shell: run `.nsh` scripts, interactive REPL, can be a default login shell (optional)
- Readable: keywords over symbols, `if x > 5 then ... end`, no `$?`, no quoting traps
- Direct system access: unknown commands resolve from PATH (`ls -la`, `git status` just work)
- A real language: variables, expressions, conditionals, loops, functions, modules, error handling
- Eventually: structured data (JSON native), pipelines (text + structured), parallel execution,
  retries/timeouts, task dependency graphs, cross-platform file/net APIs

## What Nesh Is NOT

- Not a Bash clone or a prettier terminal
- Not a wrapper collection around existing commands
- Not a replacement for general-purpose languages
- Not breaking the Unix composability model — pipelines stay sacred

## Non-Negotiable Principles

1. **Readability beats cleverness** — no cryptic symbols, no context-dependent parsing surprises
2. **Predictable grammar** — an LLM must be able to generate valid Nesh every time; that means
   minimal exceptions, one obvious way to do things
3. **Adoption on their terms** — a single `.nsh` file works with zero shell change; nobody is forced to switch
4. **Performance honesty** — beat Bash where it matters (≥70% of benchmark categories), never
   sacrifice readability for a microsecond
5. **Errors that teach** — every error carries line/column and says what was expected
6. **Modular by construction** — one-way dependency DAG, OS access only through
   `internal/shell` interfaces, builtins are plugins. MODULES.md is law.

## Success = Done When

- [x] M1: `nesh script.nsh` executes print/let/arithmetic — **done: v0.1.0 (2026-08-23)**
- [x] M2: if/else, functions, PATH commands, working REPL — **done: v0.2.0 (2026-08-23)**
- [x] M3: Helpful errors, history/multi-line REPL, JSON AST export (`--json`), file/string/math builtins, basic modules — **done: v0.3.0 (2026-08-23)**
- [~] M4: `nesh pkg` installs/runs packages, macOS/Linux/Windows builds, ≥70% benchmark wins (4/7 today), real-world scripts work (3 dogfooded) — **target Oct 19, 2026**
- [ ] A developer and an AI agent can both read the same `deploy.nsh` and agree on what it does

## The Long Game

Command → Script → Automation → Workflow → AI-assisted execution — one language the whole way.
The bet: the interface between humans and agents should be designed intentionally, not inherited
from 1989.
