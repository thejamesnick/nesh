# Nesh Script
## Product, Philosophy, and Technical Specification

---

# 1. Vision

Nesh is a modern, human-readable, AI-native scripting language and shell designed to simplify automation, system interaction, and developer workflows.

It aims to combine the simplicity of traditional shells with the structure and clarity of modern programming languages, while remaining extensible, consistent, and suitable for both humans and AI agents.

---

# 2. Purpose

The purpose of Nesh is to:

- Provide a cleaner alternative to traditional shell scripting
- Enable readable, maintainable automation scripts
- Support AI-assisted code generation and execution
- Reduce ambiguity in command parsing
- Offer a unified interface for interacting with the system

Nesh is not just a shell—it is a scripting runtime designed for both manual use and programmatic orchestration.

---

# 3. Design Principles

## 3.1 Human Readability
- Syntax should be intuitive and easy to understand
- Minimal cryptic symbols
- Clear structure over compactness

## 3.2 Consistency
- Uniform syntax across commands
- Predictable behavior
- Minimal exceptions to rules

## 3.3 AI-Native Design
- Structured parsing that can be interpreted reliably
- Commands map cleanly to semantic actions
- Easy to generate and validate programmatically

## 3.4 Simplicity with Power
- Simple for basic tasks
- Scalable for complex workflows
- Avoid unnecessary verbosity while maintaining clarity

---

# 4. Core Language Concepts

## 4.1 Basic Syntax

Example:

```
print hello
print "hello world"
```

- Commands are expressed as verbs
- Arguments follow the command
- Quotes are used only when necessary

## 4.2 Commands

Commands are the primary building blocks of Nesh. Built-in commands like `print` and `let` are handled natively. Any other command is treated as a system call — Nesh looks it up on the PATH and executes it directly. No wrapper needed.

Examples:

```nesh
print "hello world"
ls -la
git status
docker ps
```

### System commands (implemented, Phase 2)

- A bare `word ...` line runs the named program from PATH with literal
  word arguments. Flags merge naturally: `-la`, `--force`.
- Words are literals — strings, numbers, flags. No variable expansion yet
  (Phase 3 decision); expressions don't belong in command position.
- Exit codes: `run <command>` is an expression evaluating to the exit code:

  ```nesh
  let code = run git status
  if code == 0 then
    print "clean tree"
  end
  ```

- A bare command's output streams straight through to stdout/stderr.
- If a variable shadows a command name, calling it as a command is an
  error with a hint ("did you mean print x?").

## 4.3 Variables

```
let name = "John"
print name
```

## 4.4 Comparisons and Booleans

Comparison operators: `== != < > <= >=`
Boolean operators: `and or not` (short-circuiting)

Precedence (loosest to tightest): `or` → `and` → comparisons → `+ -` → `* /`

```
let x = 10
if x > 5 and x < 20 then
  print "in range"
end
print not (x == 10)   # false
```

### Truthiness (the rule, all of it)

Wherever a value is used as a condition (`if`, `while`, `and`, `or`, `not`):

| Value          | Truthy? |
|----------------|---------|
| `false`        | false   |
| `0`            | false   |
| `0.0`          | false   |
| `""` (empty)   | false   |
| everything else| true    |

No silent coercion: `"false"` is truthy (non-empty string), and comparing
values of different types with `< > <= >=` is a runtime error.

## 4.5 Control Flow

```
if balance > 100 then
  print "ok"
elif balance > 0 then
  print "low"
else
  print "empty"
end
```

Decision: Nesh has `elif` (not `else if`). Flat chains beat deep nesting —
for humans reading them and for agents generating them. Indentation inside
blocks is optional style, not syntax; blocks are closed by keywords.

Loops:

```
while i < 10
  print i
  let i = i + 1
end
```

There is no `break`/`continue` yet (Phase 3 candidate if needed). An
infinite loop is the author's responsibility, same as bash — no guard.
`for x in list` arrives with list values.

## 4.6 Functions

```
fn greet(name)
  return "hello " + name
end

print greet("world")     # call as expression
deploy("prod")           # call as statement (side effects)
```

### Scoping rules (the whole story)

1. Function **parameters** are locals of each call.
2. A `let` inside a function creates a **local**; outside, a global.
3. Reads fall back to outer scopes (locals → globals).
4. Redefining a global from inside a function is impossible with `let` —
   use explicit returns to hand values back.
5. Functions themselves live in the **global** scope.
6. A function without an explicit `return` yields `false`.
7. Recursion works (each call gets a fresh scope).

Arity is strict: calling with the wrong number of arguments is a runtime
error at the call site.

## 4.7 Standard Library

Builtins are called like functions and live in the global scope (user
code may shadow them with `let`).

Strings & lists:
```
len("hello")            # 5 — strings by runes, lists by elements
upper(s) lower(s)
split("a,b", ",")       # list: ["a", "b"]
join(list, "-")
contains(s, sub)        # true/false
```

Math:
```
abs(-7)  floor(3.9)  round(2.5)   # ints
min(a, b, ...) max(a, b, ...)
```

Files (OS access via the shell.FileSystem seam):
```
read("f.txt")           # string; runtime error if unreadable
write("f.txt", data)    # true on success
exists("f.txt")         # true/false
```

Lists & for-in:
```
let xs = split("a b c", " ")
for x in xs
  print x
end
```
The loop variable binds in the surrounding scope (same rule as `let`),
so accumulators behave exactly like in `while`.

---

# 5. Execution Model

Nesh operates as a runtime that:

1. Parses scripts into structured tokens
2. Converts commands into an internal representation
3. Executes commands sequentially or conditionally
4. Interfaces with the underlying operating system for system-level operations

Execution pipeline:

```
Source Code → Lexer → Parser → AST / IR → Executor → OS/System Calls
```

---

# 6. Technology Stack

## 6.1 Core Language Runtime
- Go (primary implementation language)
  - Fast compilation
  - Strong concurrency support
  - Cross-platform binaries
  - Suitable for CLI and system tools

## 6.2 Parser
- Custom parser written in Go
- Tokenizer + grammar-based parsing
- Support for expressions, commands, and control flow

## 6.3 CLI Interface
- Terminal-based REPL
- Script execution from files
- Interactive shell mode

## 6.4 System Integration
- OS process execution (exec calls)
- Environment variable access
- File system operations

## 6.5 Optional AI Integration Layer
- Structured command representation (JSON-like AST)
- API hooks for AI agents
- Natural language to Nesh translation layer (future)

---

# 7. Core Components to Build

## 7.1 Lexer
- Converts raw input into tokens
- Handles keywords, identifiers, strings, operators

## 7.2 Parser
- Builds structured representation (AST or IR)
- Validates syntax rules

## 7.3 Runtime / Executor
- Executes parsed instructions
- Handles built-in commands (print, let, if, etc.)
- Dispatches system commands

## 7.4 Standard Library
- Built-in commands:
  - print
  - let
  - if
  - fn

## 7.5 CLI Tool
- `nesh` command
- Run scripts:
  - nesh script.nsh
- Interactive shell mode

## 7.6 Package / Module System (Future)
- Reusable scripts
- Import/export functionality

---

# 8. File Format

Nesh scripts will use a dedicated extension:

```
.nsh
```

Example:

```
script.nsh
```

This distinguishes Nesh scripts from traditional shell scripts (.sh).

---

# 9. Syntax Philosophy

- Avoid unnecessary symbols
- Prefer keywords over symbols
- Maintain optional quoting for clarity
- Keep grammar predictable

Example comparison:

Bash:
```
echo hello world
```

Nesh:
```
print hello
print "hello world"
```

---

# 10. Target Use Cases

- Automation scripts
- Developer tooling
- DevOps workflows
- Local system orchestration
- AI-driven task execution
- CLI utilities

---

# 11. Future Roadmap

## Phase 1
- Core interpreter in Go
- Lexer and parser
- Basic commands (print, let, if)
- CLI execution

## Phase 2
- Functions
- Script modules
- Improved error handling
- Standard library expansion

## Phase 3
- AI integration layer
- Natural language to Nesh translation
- Structured execution APIs

## Phase 4
- Package manager
- Ecosystem tooling
- Cross-platform distribution

---

# 12. Success Criteria

Nesh is considered usable when:

- Scripts can automate real workflows reliably
- Developers can adopt it without steep learning curve
- Parsing is deterministic and predictable
- Execution integrates smoothly with the OS
- It can scale from small scripts to complex automation systems

---

# 13. Guiding Philosophy Summary

Nesh is built to bridge the gap between:

- Human-readable scripting
- System-level control
- AI-assisted automation

It prioritizes clarity, structure, and long-term usability over compatibility with legacy shell constraints.

---

End of Document

