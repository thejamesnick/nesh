package runtime

import (
	"strings"
	"testing"

	"nesh/internal/parser"
)

func tryRun(t *testing.T, src string) (string, *Error) {
	t.Helper()
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error for %q: %v", src, perr)
	}
	out := &fakeOutput{}
	rt := New(out)
	rerr := rt.Run(script)
	return out.b.String(), rerr
}

func TestFailCaughtByHandler(t *testing.T) {
	out, err := tryRun(t, "try\nfail \"db down\"\nprint \"never\"\non failure\nprint \"caught:\" failure\nend\nprint \"after\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "caught: db down\nafter\n" {
		t.Fatalf("got %q", out)
	}
}

func TestTryWithoutHandlerSwallows(t *testing.T) {
	out, err := tryRun(t, "try\nfail\nend\nprint \"survived\"\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "survived\n" {
		t.Fatalf("got %q", out)
	}
}

func TestUncaughtFailAborts(t *testing.T) {
	_, err := tryRun(t, "print \"a\"\nfail \"boom\"\nprint \"b\"\n")
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if err.Error() != "2:1: uncaught failure: boom" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestFailInsideFunctionIsCatchable(t *testing.T) {
	src := "fn deploy(env)\nif env == \"prod\" then\nfail \"no prod today\"\nend\nreturn true\nend\ntry\ndeploy(\"prod\")\non failure\nprint \"blocked:\" failure\nend\n"
	out, err := tryRun(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "blocked: no prod today\n" {
		t.Fatalf("got %q", out)
	}
}

func TestNestedTryInnerCatchesFirst(t *testing.T) {
	src := "try\ntry\nfail \"inner\"\non failure\nprint \"inner caught\" failure\nend\non failure\nprint \"outer caught\"\nend\n"
	out, err := tryRun(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "inner caught inner\n" {
		t.Fatalf("got %q", out)
	}
}

func TestFailInHandlerPropagatesToOuterTry(t *testing.T) {
	src := "try\ntry\nfail \"one\"\non failure\nfail failure + \"!\"\nend\non failure\nprint \"outer got:\" failure\nend\n"
	out, err := tryRun(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "outer got: one!\n" {
		t.Fatalf("got %q", out)
	}
}

func TestHardErrorsNotCatchable(t *testing.T) {
	// division by zero is a bug: it must abort even inside try
	_, err := tryRun(t, "try\nlet x = 1 / 0\non failure\nprint \"should not run\"\nend\n")
	if err == nil || !strings.Contains(err.Msg, "division by zero") {
		t.Fatalf("got %v, want hard division-by-zero abort", err)
	}
}

func TestReturnPassesThroughTry(t *testing.T) {
	src := "fn f()\ntry\nreturn 7\nend\nreturn 0\nend\nprint f()\n"
	out, err := tryRun(t, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "7\n" {
		t.Fatalf("got %q", out)
	}
}

func TestBreakPassesThroughTry(t *testing.T) {
	src := "for x in items()\ntry\nbreak\nend\nend\nprint \"done\"\n"
	script, _ := parser.Parse(src)
	out := &fakeOutput{}
	rt := New(out)
	rt.Define("items", func(args []Value) (Value, error) { return List{Int(1), Int(2)}, nil })
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if out.b.String() != "done\n" {
		t.Fatalf("got %q", out.b.String())
	}
}

func TestConditionalFailWithExitCode(t *testing.T) {
	// the realistic idiom: run a command, decide, fail
	fr := &fakeRunner{code: 1}
	rt := New(&fakeOutput{})
	rt.SetRunner(fr)
	src := "try\nlet code = run tests\nif code != 0 then\nfail \"tests failed\"\nend\nprint \"all green\"\non failure\nprint \"handler:\" failure\nend\n"
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if got := rt.out.(*fakeOutput).b.String(); got != "ran: tests\nhandler: tests failed\n" {
		t.Fatalf("got %q", got)
	}
}

func TestExitStatement(t *testing.T) {
	// exit carries a code through every layer as a plain error
	script, _ := parser.Parse("for x in items()\nexit 3\nend\n")
	rt := New(&fakeOutput{})
	rt.Define("items", func(args []Value) (Value, error) { return List{Int(1)}, nil })
	err := rt.Run(script)
	if err == nil || err.ExitCode != 3 {
		t.Fatalf("got %v, want ExitCode 3", err)
	}

	// bare exit = 0
	out, err := tryRun(t, "print \"bye\"\nexit\nprint \"never\"\n")
	if err.ExitCode != 0 {
		t.Fatalf("bare exit: got %v", err)
	}
	if out != "bye\n" {
		t.Fatalf("output before exit: %q", out)
	}

	// exit is NOT catchable by try
	_, err = tryRun(t, "try\nexit 5\non failure\nprint \"caught\"\nend\n")
	if err == nil || err.ExitCode != 5 {
		t.Fatalf("try caught exit: %v", err)
	}
}
