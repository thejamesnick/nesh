package runtime

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"nesh/internal/parser"
	"nesh/internal/shell"
)

// runCapture parses src, executes it with fr as the CommandRunner, and
// returns the output-sink contents plus the runtime error, if any.
// Note: Run happens BEFORE reading the buffer — Go evaluates return
// expressions left-to-right, so reading first would always see "".
func runCapture(t *testing.T, src string, fr *fakeRunner) (string, *Error) {
	t.Helper()
	out := &fakeOutput{}
	rt := New(out)
	rt.SetRunner(fr)
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error for %q: %v", src, perr)
	}
	err := rt.Run(script)
	return out.b.String(), err
}

func TestCaptureOutput(t *testing.T) {
	fr := &fakeRunner{fn: func(name string, args []string, stdin string, stdout io.Writer) int {
		io.WriteString(stdout, "hello\n")
		return 0
	}}
	out, err := runCapture(t, "let x = capture gen\nprint \"got:\" x\n", fr)
	if err != nil {
		t.Fatal(err)
	}
	if out != "got: hello\n" {
		t.Fatalf("got %q", out)
	}
}

func TestCaptureStripsTrailingNewlines(t *testing.T) {
	fr := &fakeRunner{fn: func(name string, args []string, stdin string, stdout io.Writer) int {
		io.WriteString(stdout, "a\nb\n\n")
		return 0
	}}
	// interior newlines survive; trailing ones don't (bash $(...) style)
	out, err := runCapture(t, "let x = capture gen\nprint x == \"a\\nb\"\n", fr)
	if err != nil {
		t.Fatal(err)
	}
	if out != "true\n" {
		t.Fatalf("trailing newlines not stripped: got %q", out)
	}
}

func TestCaptureDiscardsExitCode(t *testing.T) {
	fr := &fakeRunner{fn: func(name string, args []string, stdin string, stdout io.Writer) int {
		io.WriteString(stdout, "output")
		return 7 // nonzero exit must neither abort nor surface through capture
	}}
	out, err := runCapture(t, "let x = capture failing\nprint \"ok:\" x\n", fr)
	if err != nil {
		t.Fatalf("nonzero exit aborted capture: %v", err)
	}
	if out != "ok: output\n" {
		t.Fatalf("got %q", out)
	}
}

func TestCapturePipeline(t *testing.T) {
	fr := &fakeRunner{fn: func(name string, args []string, stdin string, stdout io.Writer) int {
		if name == "gen" {
			io.WriteString(stdout, "x\ny\nz\n")
		} else {
			io.WriteString(stdout, fmt.Sprintf("%d\n", strings.Count(stdin, "\n")))
		}
		return 0
	}}
	out, err := runCapture(t, "let n = capture gen | wc\nprint \"lines:\" n\n", fr)
	if err != nil {
		t.Fatal(err)
	}
	if out != "lines: 3\n" {
		t.Fatalf("pipeline capture got %q", out)
	}
}

func TestCaptureRedirectWins(t *testing.T) {
	// bash-style: `capture cmd > file` — the redirect claims stdout, so
	// the captured value is empty and the file gets the output.
	opener := newFakeOpener()
	fr := &fakeRunner{fn: func(name string, args []string, stdin string, stdout io.Writer) int {
		io.WriteString(stdout, "to-file\n")
		return 0
	}}
	out := &fakeOutput{}
	rt := New(out)
	rt.SetRunner(fr)
	rt.SetFileSystem(fakeFSWithOpener{opener})
	script, perr := parser.Parse("let x = capture gen > log.txt\nprint \"captured:[\" x \"]\"\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if out.b.String() != "captured:[  ]\n" {
		t.Fatalf("capture should be empty when redirected, got %q", out.b.String())
	}
	if got := opener.files["log.txt"].String(); got != "to-file\n" {
		t.Fatalf("file got %q, want redirect to win", got)
	}
}

func TestCaptureCommandBuiltin(t *testing.T) {
	// command builtins (printf/echo-class) capture without a CommandRunner
	out := &fakeOutput{}
	rt := New(out)
	rt.DefineCommand("printf", func(args []string, stdout io.Writer) int {
		io.WriteString(stdout, args[0]+"\n")
		return 0
	})
	script, perr := parser.Parse("let x = capture printf hi\nprint x\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if out.b.String() != "hi\n" {
		t.Fatalf("builtin capture got %q", out.b.String())
	}
}

func TestCaptureWithoutRunner(t *testing.T) {
	_, err := run(t, "let x = capture git status\n")
	if err == nil || err.Msg != "system commands are not available here" {
		t.Fatalf("got %v, want no-runner error", err)
	}
}

func TestCaptureStderrStillPrints(t *testing.T) {
	// End-to-end with a real spawn: stdout is captured into the variable;
	// stderr still streams to the output sink.
	out := &fakeOutput{}
	rt := New(out)
	rt.SetRunner(shell.RealRunner{})
	script, perr := parser.Parse("let x = capture sh -c \"echo out; echo err >&2\"\nprint \"got:\" x\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if got := out.b.String(); got != "err\ngot: out\n" {
		t.Fatalf("stderr should print, stdout should be captured: got %q", got)
	}
}
