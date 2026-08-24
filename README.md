# Nesh

> A modern, human-readable shell and scripting language built for the agent era.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8.svg)](https://golang.org/)
[![Status: In Development](https://img.shields.io/badge/status-in%20development-orange.svg)]()

---

## What is Nesh?

Nesh is a shell and scripting language designed to be clean, predictable, and easy to work with — for both humans and AI agents.

It does the same job as Bash. You can use it to write automation scripts, run it interactively as your terminal shell, or set it as your default shell on macOS and Linux. The difference is how it feels to read and write.

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

No cryptic symbols. No quoting landmines. Just readable code that does what it says.

---

## Why Nesh?

We're in the agent era. AI tools are already writing and running shell scripts on behalf of developers. The problem is they're doing it in Bash — a 35-year-old language never designed for this.

Nesh is built for the world we're actually in:

- Humans can read and write it without a manual
- AI agents can generate it reliably without ambiguity
- Anyone can review AI-generated Nesh scripts and understand them instantly

Read the full story in [docs/WHY_NESH.md](docs/WHY_NESH.md).

---

## Usage

```bash
# Run a script
nesh script.nsh

# Open interactive shell (REPL)
nesh

# Run a one-liner
nesh -c "print hello"

# Agent API: JSON AST + structured execution events
nesh --json script.nsh
```

You can also set Nesh as your default shell on macOS or Linux — but you don't have to. Start with a single `.nsh` file. No pressure, no forced migration.

---

## Syntax at a Glance

| Concept       | Nesh                          | Bash                          |
|---------------|-------------------------------|-------------------------------|
| Print         | `print "hello"`               | `echo "hello"`                |
| Variable      | `let x = 5`                   | `x=5`                         |
| Condition     | `if x > 5 then ... end`       | `if [ "$x" -gt 5 ]; then ... fi` |
| Function      | `fn greet(name) ... end`      | `greet() { ... }`             |
| Loops         | `while i < 10 ... end` / `for x in list ... end` | `while`, `for` |
| System command| `ls -la`                      | `ls -la`                      |
| Exit code     | `let code = run git status`   | `$?`                          |
| Modules       | `import "utils.nsh" as u`     | `source utils.sh`             |
| Stdlib        | `split(s, ",")`, `len(x)`, `read(f)` | awk/sed/grep gymnastics |

---

## Tech Stack

- Language: Go
- File extension: `.nsh`
- Execution: `Source → Lexer → Parser → AST → Executor → OS`

---

## Project Status

Phase 3 (AI & Polish) complete. Nesh runs real automation scripts:

- Full language: arithmetic, booleans, if/elif, functions + recursion, while/for-in loops
- System commands straight from PATH with exit-code capture via `run`
- Standard library: string ops, math, file read/write
- Modules: `import "x.nsh"` with optional namespacing (`as u`) and cycle detection
- REPL with multi-line blocks, history, and line editing
- `nesh --json`: machine-readable AST + execution events for AI agents
- Faster than bash in every measured category (see benchmarks/results/)

Roadmap:
- Phase 4 — Package manager, cross-platform distribution, v1.0 polish

---

## Docs

| File | Description |
|------|-------------|
| [docs/WHY_NESH.md](docs/WHY_NESH.md) | The philosophy and reason Nesh exists |
| [docs/CAPABILITIES.md](docs/CAPABILITIES.md) | What Nesh can do right now (v0.3.0 snapshot) |
| [docs/LIMITATIONS.md](docs/LIMITATIONS.md) | What Nesh can't do yet, and where it's tracked |
| [docs/TECHNICAL_SPEC.md](docs/TECHNICAL_SPEC.md) | Full product and technical specification |
| [docs/TECH_STACK.md](docs/TECH_STACK.md) | Technology decisions and component breakdown |
| [docs/BENCHMARKS.md](docs/BENCHMARKS.md) | Performance strategy and benchmarking plan |
| [docs/PERFORMANCE.md](docs/PERFORMANCE.md) | Go vs Rust vs C analysis and optimization notes |
| [docs/PROJECT_SUMMARY.md](docs/PROJECT_SUMMARY.md) | Action plan and phase breakdown |

---

## Contributing

Open source from day one. Contribution guidelines coming as the project takes shape. Watch this space.

---

## License

MIT — see [LICENSE](LICENSE) for details.
