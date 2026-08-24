# Archived

> Historical note from before the project docs existed (2026-08-23). Superseded by [goal.md](../../goal.md), [task.md](../../task.md) and the files in [docs/](..). Kept as-is for context.

Hahaha 😂 let's pretend **Nesh is finished** and we're actually sitting in your terminal.

The goal is that it should feel familiar enough that someone coming from Bash can start immediately, but **much easier to read**.

## 🐚 Welcome to Nesh

You open Terminal:

```text
$ nesh

Nesh 1.0.0
A human-readable shell for automation and scripting.

nesh>
```

Now let's actually use it.

---

## 1. Basic commands

You don't need to learn some weird syntax just to run programs.

```text
nesh> run npm install
nesh> run npm test
nesh> run npm run build
```

Or simply:

```text
nesh> npm install
nesh> npm test
nesh> npm run build
```

Nesh understands that these are external commands.

---

# 2. Variables

Instead of Bash:

```bash
name="Nick"
echo "$name"
```

Nesh:

```text
let name = "Nick"

say name
```

Output:

```text
Nick
```

You can work with actual values:

```text
let users = 100
let price = 5

let revenue = users * price

say revenue
```

Output:

```text
500
```

---

# 3. Conditions

Traditional shell:

```bash
if [ "$ENV" = "production" ]; then
    ...
fi
```

Nesh:

```text
let environment = "production"

if environment == "production":
    say "Deploying production"
else:
    say "Development environment"
```

Much easier to read.

---

# 4. Functions

```text
function build:
    say "Building..."
    run npm run build

build
```

You could eventually have arguments:

```text
function greet name:
    say "Hello " + name

greet "Nick"
```

---

# 5. Files

Instead of constantly reaching for Unix utilities:

```text
nesh> read package.json
nesh> write "hello.txt" "Hello world"
nesh> copy "app.json" "backup/app.json"
nesh> move "old.txt" "archive/old.txt"
```

The runtime understands these operations directly.

---

# 6. Structured data 🔥

This is where Nesh starts getting interesting.

Imagine:

```text
let users = fetch "https://api.example.com/users"

for user in users:
    say user.name
```

No:

```text
curl
jq
grep
awk
sed
```

just to process an API response.

Nesh understands the data.

---

# 7. Pipelines

But you **still keep Unix's beautiful pipeline idea**.

```text
run git log
| filter "fix"
| take 10
```

You can also eventually have structured pipelines:

```text
users
| filter user.active
| map user.name
| take 10
```

That's much more powerful than treating everything as text.

---

# 8. Parallel work

Imagine you're working on a full-stack project.

Instead of opening three terminals:

```text
nesh> parallel:
    npm run frontend
    npm run backend
    npm run worker
```

Nesh manages the processes.

You could get:

```text
[frontend] started
[backend] started
[worker] started

[backend] ready
[worker] ready
[frontend] ready

All processes running.
```

---

# 9. Error handling

Something fails?

Instead of Bash's sometimes-confusing exit-code behavior:

```text
nesh> try:
    run npm test

on failure:
    say "Tests failed"
```

And Nesh knows the difference between:

```text
success
failure
timeout
cancelled
crashed
```

---

# 10. Retry

For unreliable operations:

```text
retry 3:
    run deploy
```

Nesh can report:

```text
Attempt 1... failed
Attempt 2... failed
Attempt 3... success

Deployment complete.
```

---

# 11. Environment

```text
environment:
    NODE_ENV = "production"
    API_URL = secret("API_URL")

run npm start
```

Nesh manages the environment for the process.

---

# 12. Now let's do something REAL 😂

Imagine we've just cloned your new product.

```text
nesh> git clone github.com/nick/my-product
nesh> cd my-product
```

Then:

```text
nesh> install dependencies
```

Nesh detects the project:

```text
Detected Node.js project.

Installing dependencies...
✓ package.json found
✓ npm detected

Dependencies installed.
```

Then:

```text
nesh> test
```

Output:

```text
Running tests...

✓ 48 passed
✓ 0 failed

Tests successful.
```

Then:

```text
nesh> build
```

Output:

```text
Building application...

✓ Type checking
✓ Bundling
✓ Assets generated

Build successful.
```

Then:

```text
nesh> deploy
```

And Nesh can execute the project's configured deployment workflow.

---

# And THIS is where your original Codex idea comes back 👀

Imagine Nesh has a first-class automation concept:

```text
goal "Ship authentication"
```

Nesh could understand a workflow like:

```text
goal "Ship authentication":

    inspect project

    implement authentication

    test

    if tests.failed:
        diagnose
        fix
        retry

    build

    verify

    report
```

Now you've got:

```text
                 NESH
                  │
        ┌─────────┴─────────┐
        │                   │
   HUMAN USER           AI AGENT
        │                   │
        └─────────┬─────────┘
                  ↓
          SAME AUTOMATION
             RUNTIME
```

That's the part I find genuinely interesting.

You're not necessarily trying to make **another terminal**.

You're building a **human-readable execution language** that can eventually be understood by both developers and agents.

---

## And the `.nsh` file would look beautiful

Imagine opening:

```text
deploy.nsh
```

and seeing:

```text
environment:
    NODE_ENV = "production"

task test:
    run npm test

task build:
    depends_on test
    run npm run build

task deploy:
    depends_on build

    try:
        run ./scripts/deploy
    on failure:
        retry 3

task ship:
    depends_on deploy

    say "🚀 Product shipped!"
```

A developer can read that almost like English.

The computer gets a precise execution graph.

And an AI agent can inspect it and understand what the workflow is trying to accomplish.

**That is the Nesh I'd want to build.** Not "Bash but with nicer `if` statements" 😂 — **a modern shell/runtime where commands, programs, workflows, structured data, and automation are first-class concepts.**
