# Nesh — Project Understanding Summary

*Compiled from all project docs: README.md, About.txt, PLAN.txt, Brand.txt, Look.txt,
"how it might be haha.txt", and everything in docs/.*

---

## What Nesh Is

**Nesh** is a modern, human-readable shell and scripting language built for the agent era.
It does the same job as Bash — automation scripts, interactive terminal use, even a default
shell on macOS/Linux — but with clean, predictable syntax that humans can read without a
manual and AI agents can generate reliably without ambiguity.

- **Language name:** Nesh
- **Runtime / CLI binary:** `nesh`
- **Script extension:** `.nsh`
- **License:** MIT
- **Status:** Planning complete, zero code written yet (greenfield)

Naming identity: **N**ew + **sh**ell → Nesh, with `.nsh` feeling familiar to anyone who knows `.sh`.

---

## Why It Exists (The Core Bet)

AI agents increasingly write and execute shell scripts — in Bash, a 35-year-old language
never designed for machine generation or unambiguous parsing. Nesh's core insight:

> **The same properties that make code readable for humans make it reliable for AI.**

Clean grammar, keywords over symbols, no quoting landmines, predictable structure.
A human can review an agent-generated Nesh script in seconds.

Adoption philosophy: nobody is forced to switch shells. Start with a single `.nsh` file;
make it your default shell only if you want to.

---

## The Language

```nesh
let name = "world"
print "hello" name

if balance > 100 then
  print "looking good"
end

fn deploy(env)
  docker compose up -d
  print "deployed to" env
end
```

Key semantics:
- **Commands are verbs**; built-ins (`print`, `let`, `if`, `fn`) handled natively
- **Unknown commands resolve from PATH directly** — `ls -la`, `git status` just work, no `run` wrapper
- Real programming constructs: variables, expressions, conditionals, loops, functions, modules, error handling
- Eventual first-class: structured data (JSON natively, no jq), pipelines (text and structured),
  parallel execution, retries/timeouts, task dependency graphs, environment/secrets management,
  cross-platform filesystem and networking APIs
- `try: ... on failure:` style structured errors instead of `$?`

Long-term vision: a shared automation runtime both humans and AI agents read and execute —
commands → scripts → workflows → AI-assisted execution, one language the whole way.

---

## Architecture & Tech Stack

- **Implementation language: Go** (decision made — note the earliest vision doc said Rust, but all current docs settle on Go)
- Pipeline: `Source → Lexer → Parser → AST/IR → Executor → OS`
- Planned structure: `cmd/nesh/` (CLI), `internal/{lexer,parser,runtime,builtin}/`, `benchmarks/`, `testdata/`

**Performance strategy:** Start in Go for velocity; a well-written Go interpreter beats Bash on
anything non-trivial. Shell bottlenecks are usually OS calls (process spawn, file I/O), which are
language-agnostic. If profiling later shows interpreter hot spots, rewrite only those paths in
Rust or C via cgo. Targets: ≥70% benchmark wins vs Bash, ≤2x Bash startup, ≥2x Bash on
loop/variable-heavy work.

---

## Roadmap

| Phase | Weeks | Goal | Done when |
|-------|-------|------|-----------|
| 1 — Foundation | 1–2 | Lexer, parser, executor, CLI, print/let | `nesh script.nsh` runs print/let basics |
| 2 — Core Features | 3–4 | if/else, functions, PATH command dispatch | Scripts with control flow + functions work |
| 3 — AI & Polish | 5–6 | Good errors (line/col), REPL (history, multi-line), JSON AST export, stdlib (files, math, strings), modules | REPL usable, structured output valid |
| 4 — Ecosystem | 7–8 | Package manager (`nesh pkg`), cross-platform builds, optimization, docs | Installable and useful for real automation |

Each phase requires: features + unit tests (>85–90% pass) + integration tests + manual testing +
benchmarks + updated docs.

---

## Immediate Next Actions (from PLAN.txt)

1. `go mod init nesh`
2. Create directory structure (`cmd/nesh`, `internal/lexer`, etc.)
3. Write first lexer test
4. Implement lexer for basic tokens

---

## What Nesh Is NOT

Not a Bash clone, not a prettier terminal, not a command alias system, not a replacement for
every language. It's a better model for human-readable system automation — one runtime where
a human and an AI agent can look at the same `deploy.nsh` and both understand exactly what it does.
