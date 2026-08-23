// Command nesh is the Nesh shell: script runner, one-liner runner, and REPL.
//
// Contract (MODULES.md): composition root only — builds the object graph
// (stdout → runtime ← parser) and gets out of the way. Language semantics
// live in internal/*.
package main

import (
	"bufio"
	"fmt"
	"os"

	"nesh/internal/parser"
	"nesh/internal/runtime"
	"nesh/internal/shell"
)

const version = "0.1.0"

const usage = `Nesh ` + version + ` — a human-readable shell for automation and scripting.

Usage:
  nesh script.nsh      run a script file
  nesh -c "print 1"    run a one-liner
  nesh                 interactive shell (REPL)
  nesh -h | help       this message

Phase 1 language: let, print, arithmetic (+ - * /), strings, ints, floats.
Everything else is on the roadmap (task.md).`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	switch {
	case len(args) == 0:
		return repl(os.Stdin, os.Stdout)
	case args[0] == "-h" || args[0] == "help" || args[0] == "--help":
		fmt.Print(usage)
		return 0
	case args[0] == "-c":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, `usage: nesh -c "print 1"`)
			return 2
		}
		return execSource(args[1])
	default:
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: nesh script.nsh  (single file for now)")
			return 2
		}
		return execFile(args[0])
	}
}

func execFile(path string) int {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nesh: %v\n", err)
		return 2
	}
	return execSource(string(src))
}

func execSource(src string) int {
	script, perr := parser.Parse(src)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", perr)
		return 1
	}
	out := bufio.NewWriter(os.Stdout)
	rt := runtime.New(out)
	rt.SetRunner(shell.RealRunner{})
	if rerr := rt.Run(script); rerr != nil {
		out.Flush()
		fmt.Fprintf(os.Stderr, "error: %v\n", rerr)
		return 1
	}
	out.Flush()
	return 0
}

// repl reads lines, evaluates each in a shared runtime (globals persist),
// and keeps going after errors. History/editing arrive in Phase 3.
func repl(in *os.File, outFile *os.File) int {
	w := bufio.NewWriter(outFile)
	defer w.Flush()
	interactive := isTerminal(in)
	if interactive {
		fmt.Fprintf(w, "Nesh %s\n", version)
	}
	rt := runtime.New(w)
	rt.SetRunner(shell.RealRunner{})
	sc := bufio.NewScanner(in)
	for {
		if interactive {
			fmt.Fprint(w, "nesh> ")
			w.Flush()
		}
		if !sc.Scan() {
			fmt.Fprintln(w)
			return 0
		}
		line := sc.Text()
		if line == "exit" {
			return 0
		}
		script, perr := parser.Parse(line)
		if perr != nil {
			fmt.Fprintf(w, "parse error: %v\n", perr)
			w.Flush()
			continue
		}
		if rerr := rt.Run(script); rerr != nil {
			fmt.Fprintf(w, "error: %v\n", rerr)
		}
		w.Flush()
	}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
