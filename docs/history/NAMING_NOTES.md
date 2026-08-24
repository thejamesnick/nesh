# Archived

> Historical note from before the project docs existed (2026-08-23). Superseded by [goal.md](../../goal.md), [task.md](../../task.md) and the files in [docs/](..). Kept as-is for context.

😂 **Nesh / `nsh` is actually pretty damn good.**

It has the right relationship to `sh` without sounding like you're literally cloning it.

I'd structure it like this:

```text
Nesh
│
├── Language: Nesh
├── Shell: Nesh Shell
├── Runtime / CLI: nesh
└── Script extension: .nsh
```

So you get:

```text
build.nsh
deploy.nsh
setup.nsh
agent.nsh
```

And execute them with:

```bash
nesh build.nsh
```

Interactive shell:

```text
$ nesh

nesh> say "Hello"
Hello

nesh> run npm test
...
```

That is **clean**.

---

## Why Nesh works

The name has a subtle connection to:

**N**ew + **sh**ell → **Nesh**

And `.nsh` immediately feels familiar to someone who understands `.sh`.

But the important distinction is:

> **Nesh isn't "new Bash."**

It's a new shell model.

You could brand it:

# **Nesh**

### *A modern human-readable shell for automation and scripting.*

Or eventually:

> **Nesh — The human-readable shell.**

---

## The naming hierarchy is clean

```text
                 NESH
                  │
       ┌──────────┴──────────┐
       │                     │
    Language               Runtime
       │                     │
     Nesh                  nesh
       │
       ▼
    .nsh files
```

Example:

```text
project/
├── setup.nsh
├── dev.nsh
├── test.nsh
├── build.nsh
└── deploy.nsh
```

Then:

```text
nesh setup.nsh
nesh dev.nsh
nesh test.nsh
nesh build.nsh
nesh deploy.nsh
```

And interactive:

```text
nesh>
```

That's a **real language identity**.

---

### One thing I'd do before locking it 😂

Because **Nesh** is a name that could already exist in software/products, I'd check current naming conflicts, GitHub availability, package registries, domains, and trademarks before committing.

But conceptually?

**Nesh + `nsh` + `nesh` runtime = 🔥.**

It fits the project far better than trying to force something like `.flow` if what you want is to establish a recognizable **new shell family**.
