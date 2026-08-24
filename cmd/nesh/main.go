// Command nesh is the Nesh shell: script runner, one-liner runner, and REPL.
//
// Contract (MODULES.md): composition root only — builds the object graph
// (stdout → runtime ← parser) and gets out of the way. Language semantics
// live in internal/*.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"nesh/internal/ast"
	"nesh/internal/builtin"
	"nesh/internal/parser"
	"nesh/internal/runtime"
	"nesh/internal/shell"
)

const version = "0.3.0-dev"

const usage = `Nesh ` + version + ` — a human-readable shell for automation and scripting.

Usage:
  nesh script.nsh      run a script file
  nesh -c "print 1"    run a one-liner
  nesh                 interactive shell (REPL)
  nesh --json file     JSON AST + execution events (agent API)
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
	case args[0] == "--json":
		rest := args[1:]
		switch {
		case len(rest) == 1:
			return execFileJSON(rest[0])
		case len(rest) == 2 && rest[0] == "-c":
			return execSourceJSON(rest[1])
		default:
			fmt.Fprintln(os.Stderr, `usage: nesh --json script.nsh | nesh --json -c "print 1"`)
			return 2
		}
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
	return execSourceIn(string(src), filepath.Dir(path))
}

// jsonReport is the agent-API document emitted by nesh --json.
type jsonReport struct {
	Nesh       string          `json:"nesh"`
	Status     string          `json:"status"` // "ok" | "parse_error" | "runtime_error" | "io_error"
	Ast        *ast.Script     `json:"ast,omitempty"`
	ParseError *jsonErrInfo    `json:"parse_error,omitempty"`
	RuntimeErr *jsonErrInfo    `json:"runtime_error,omitempty"`
	Events     []runtime.Event `json:"events,omitempty"`
}

type jsonErrInfo struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func emitJSON(report jsonReport) int {
	report.Nesh = version
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil { // unreachable for our types
		fmt.Fprintf(os.Stderr, "nesh: json encoding failed: %v\n", err)
		return 1
	}
	fmt.Println(string(b))
	switch report.Status {
	case "ok":
		return 0
	case "io_error":
		return 2
	default:
		return 1
	}
}

func runWithEvents(src, baseDir string) (*ast.Script, []runtime.Event, *parser.Error, *runtime.Error) {
	script, perr := parser.Parse(src)
	if perr != nil {
		return nil, nil, perr, nil
	}
	var events []runtime.Event
	rt := runtime.New(bufio.NewWriter(io.Discard))
	rt.SetRunner(shell.RealRunner{})
	rt.SetStdin(os.Stdin)
	rt.SetBaseDir(baseDir)
	builtin.RegisterAll(rt, shell.RealFS{})
	rt.SetFileSystem(shell.RealFS{})
	rt.SetRuntimeFactory(func(child *runtime.Runtime) {
		builtin.RegisterAll(child, shell.RealFS{})
	})
	rt.SetEventSink(func(e runtime.Event) { events = append(events, e) })
	rerr := rt.Run(script)
	return script, events, nil, rerr
}

func execSourceJSON(src string) int {
	return execSourceJSONIn(src, "")
}

func execSourceJSONIn(src, baseDir string) int {
	script, events, perr, rerr := runWithEvents(src, baseDir)
	report := jsonReport{Status: "ok", Events: events, Ast: script}
	if perr != nil {
		report.Status = "parse_error"
		report.ParseError = &jsonErrInfo{perr.Msg, perr.Line, perr.Column}
		return emitJSON(report)
	}
	if rerr != nil {
		report.Status = "runtime_error"
		report.RuntimeErr = &jsonErrInfo{rerr.Msg, rerr.Line, rerr.Column}
	}
	return emitJSON(report)
}

func execFileJSON(path string) int {
	src, err := os.ReadFile(path)
	if err != nil {
		return emitJSON(jsonReport{Status: "io_error"})
	}
	return execSourceJSONIn(string(src), filepath.Dir(path))
}

func execSource(src string) int {
	return execSourceIn(src, "")
}

func execSourceIn(src, baseDir string) int {
	script, perr := parser.Parse(src)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", perr)
		return 1
	}
	out := bufio.NewWriter(os.Stdout)
	rt := runtime.New(out)
	rt.SetRunner(shell.RealRunner{})
	rt.SetStdin(os.Stdin)
	rt.SetBaseDir(baseDir)
	builtin.RegisterAll(rt, shell.RealFS{})
	rt.SetFileSystem(shell.RealFS{})
	rt.SetRuntimeFactory(func(child *runtime.Runtime) {
		builtin.RegisterAll(child, shell.RealFS{})
	})
	if rerr := rt.Run(script); rerr != nil {
		out.Flush()
		fmt.Fprintf(os.Stderr, "error: %v\n", rerr)
		return 1
	}
	out.Flush()
	return 0
}

// repl reads lines (or whole blocks, once a block keyword opens), evaluates
// each in a shared runtime (globals persist), and keeps going after errors.
func repl(in *os.File, outFile *os.File) int {
	w := bufio.NewWriter(outFile)
	defer w.Flush()
	interactive := isTerminal(in)
	if interactive {
		fmt.Fprintf(w, "Nesh %s\n", version)
		w.Flush()
	}

	var rl lineReader = newPlainReader(in)
	if interactive {
		raw, err := newRawLineReader(in, outFile)
		if err == nil {
			rl = raw
			defer rl.Close()
		} else {
			fmt.Fprintf(os.Stderr, "nesh: raw mode unavailable (%v); falling back to plain input\n", err)
		}
	}

	rt := runtime.New(w)
	rt.SetRunner(shell.RealRunner{})
	rt.SetStdin(os.Stdin)
	builtin.RegisterAll(rt, shell.RealFS{})
	rt.SetFileSystem(shell.RealFS{})
	rt.SetRuntimeFactory(func(child *runtime.Runtime) {
		builtin.RegisterAll(child, shell.RealFS{})
	})
	for {
		line, err := rl.ReadLine("nesh> ")
		if err == io.EOF || line == "exit" && err == nil {
			return 0
		}
		if err == errInterrupt {
			continue
		}
		if err != nil {
			fmt.Fprintf(w, "nesh: %v\n", err)
			return 1
		}
		if line == "" {
			continue
		}

		// Collect continuation lines until every if/fn/while block closes.
		buf := line
		for parser.OpenBlocks(buf) > 0 {
			more, err := rl.ReadLine("....> ")
			if err == io.EOF {
				return 0
			}
			if err == errInterrupt {
				buf = ""
				break
			}
			if err != nil {
				fmt.Fprintf(w, "nesh: %v\n", err)
				return 1
			}
			buf += "\n" + more
		}
		if buf == "" {
			continue
		}
		if buf == "exit" {
			return 0
		}

		script, perr := parser.Parse(buf)
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
