Real talk — Go is solid but if you want to beat Bash on benchmarks, especially raw execution speed, you need to think carefully.

Where Go wins over Bash easily:

startup time for complex scripts (Bash gets slow with logic-heavy scripts)
parsing speed — Go is compiled, Bash is interpreted line by line
concurrency — Go has goroutines, Bash has nothing close
memory management — predictable and efficient
Where it gets competitive:

for simple one-liner commands, Bash is hard to beat because it's just calling the OS directly with almost zero overhead
Nesh will also be calling the OS, so the gap closes fast once your executor is tight
The honest answer on Go vs alternatives:

Language	Speed	Dev Speed	Binary Size	System Access
Go	fast	fast	small	excellent
Rust	fastest	slow	smallest	excellent
C	fastest	very slow	smallest	excellent
Go	fast	fast	small	excellent
If raw benchmark wins are the goal, Rust would edge out Go on execution speed. But Rust is significantly harder and slower to build with.

My recommendation:

Start with Go. Get the language right, get the syntax right, get the automation power right. A well-written Go interpreter will beat Bash on anything non-trivial. If benchmarks later show bottlenecks, you can rewrite hot paths in Rust or C.

Human readability is where Nesh will truly beat Bash — no benchmark needed for that, it's just obvious when you look at the code side by side.

eah, absolutely. Go has solid interop with both C and Rust.

Go + C (cgo)

built into Go, no extra tooling
you can call C functions directly from Go
common pattern: write hot paths in C, wrap them in Go
downside: cgo adds overhead and complicates cross-compilation
Go + Rust

compile Rust to a C-compatible library (.so, .dylib, .dll)
call it from Go via cgo
gives you Rust's speed where you need it, Go's ergonomics everywhere else
Hybrid approach that makes sense:

Build the whole thing in Go first
Profile it — find the actual bottlenecks (lexer? parser? executor?)
Rewrite just those hot paths in Rust or C
Keep the CLI, REPL, and high-level logic in Go
This is how a lot of performance-critical tools work. For example, ripgrep is pure Rust, but tools like fd mix approaches.

Reality check though:

For a shell/scripting language, the bottleneck is almost never the interpreter itself — it's the OS calls (spawning processes, file I/O, etc.). Those are the same speed no matter what language you use.

So Go alone will likely be fast enough. But yeah, if you want to optimize later, you can drop in C/Rust for specific pieces without rewriting everything.

Sound good?