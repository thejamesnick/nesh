package runtime

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"nesh/internal/parser"
)

// fakeOutput captures everything print writes — no real stdout touched.
type fakeOutput struct{ b strings.Builder }

func (f *fakeOutput) WriteString(s string) (int, error) { return f.b.WriteString(s) }
func (f *fakeOutput) String() string                    { return f.b.String() }

// run executes src and returns the captured output plus the runtime error, if any.
func run(t *testing.T, src string) (string, *Error) {
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

func TestPrintLet(t *testing.T) {
	out, err := run(t, "let name = \"world\"\nprint \"hello\" name\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world\n" {
		t.Fatalf("got %q, want %q", out, "hello world\n")
	}
}

func TestArithmetic(t *testing.T) {
	cases := []struct {
		src, want string
	}{
		{"print 1 + 2\n", "3\n"},
		{"print 1 + 2 * 3\n", "7\n"},
		{"print (1 + 2) * 3\n", "9\n"},
		{"print 10 / 3\n", "3\n"},     // integer division
		{"print 10 / 4.0\n", "2.5\n"}, // float promotion
		{"print 2.5 * 2\n", "5\n"},    // float result formats minimal
		{"print -5 + 3\n", "-2\n"},    // unary minus
		{"print 1.5 + 1.25\n", "2.75\n"},
		{"let x = 4\nprint x * x\n", "16\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestStringConcat(t *testing.T) {
	out, err := run(t, "let a = \"foo\"\nlet b = \"bar\"\nprint a + b\n")
	if err != nil || out != "foobar\n" {
		t.Fatalf("got %q, %v", out, err)
	}
}

func TestRuntimeErrors(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"print missing\n", "1:7: undefined variable: missing"},
		{"print 1 / 0\n", "1:9: division by zero"},
		{"print 1.0 / 0\n", "1:11: division by zero"},
		{"print \"a\" + 1\n", "1:11: cannot join string and 1 with \"+\""},
	}
	for _, c := range cases {
		_, err := run(t, c.src)
		if err == nil {
			t.Errorf("%q: expected error, got none", c.src)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("%q: got %q, want %q", c.src, err.Error(), c.want)
		}
	}
}

func TestGlobalsPersistAcrossRuns(t *testing.T) {
	out := &fakeOutput{}
	rt := New(out)
	run1, _ := parser.Parse("let x = 40\n")
	if err := rt.Run(run1); err != nil {
		t.Fatal(err)
	}
	run2, _ := parser.Parse("print x + 2\n")
	if err := rt.Run(run2); err != nil {
		t.Fatal(err)
	}
	if out.String() != "42\n" {
		t.Fatalf("got %q, want %q — globals must survive for the REPL", out.String(), "42\n")
	}
}

func TestBarePrint(t *testing.T) {
	out, err := run(t, "print\n")
	if err != nil || out != "\n" {
		t.Fatalf("bare print should emit one empty line; got %q, %v", out, err)
	}
}

func TestComparisons(t *testing.T) {
	cases := []struct{ src, want string }{
		// numbers, incl. int/float promotion
		{"print 1 == 1\n", "true\n"},
		{"print 1 == 2\n", "false\n"},
		{"print 1 != 2\n", "true\n"},
		{"print 2 > 1\n", "true\n"},
		{"print 2 < 1\n", "false\n"},
		{"print 3 >= 3\n", "true\n"},
		{"print 3 <= 2\n", "false\n"},
		{"print 1 < 1.5\n", "true\n"}, // int promoted to float
		// strings
		{"print \"a\" == \"a\"\n", "true\n"},
		{"print \"a\" != \"b\"\n", "true\n"},
		{"print \"apple\" < \"banana\"\n", "true\n"},
		// booleans
		{"print (1 < 2) == true\n", "true\n"},
		// precedence: comparison binds looser than arithmetic
		{"print 1 + 2 == 3\n", "true\n"},
		{"print 2 * 3 >= 5 + 1\n", "true\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestBooleanOperators(t *testing.T) {
	cases := []struct{ src, want string }{
		{"print true and false\n", "false\n"},
		{"print true and true\n", "true\n"},
		{"print false or true\n", "true\n"},
		{"print not true\n", "false\n"},
		{"print not 0\n", "true\n"}, // truthiness: 0 is false
		{"print not \"\"\n", "true\n"},
		// and binds tighter than or
		{"print true or false and false\n", "true\n"},
		{"print (true or false) and false\n", "false\n"},
		// comparisons feed straight into boolean ops
		{"let x = 10\nprint x > 5 and x < 20\n", "true\n"},
		{"let x = 10\nprint x == 10 or x == 99\n", "true\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestTruthiness(t *testing.T) {
	cases := []struct{ src, want string }{
		{"print not 0\n", "true\n"},      // Int(0) falsy
		{"print not 1\n", "false\n"},     // nonzero truthy
		{"print not 0.0\n", "true\n"},    // Float(0.0) falsy
		{"print not 0.5\n", "false\n"},   //
		{"print not \"\"\n", "true\n"},   // empty string falsy
		{"print not \"x\"\n", "false\n"}, // non-empty truthy
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestIfElse(t *testing.T) {
	cases := []struct{ src, want string }{
		// then branch
		{"let x = 10\nif x > 5 then\nprint \"big\"\nend\n", "big\n"},
		// else branch
		{"let x = 3\nif x > 5 then\nprint \"big\"\nelse\nprint \"small\"\nend\n", "small\n"},
		// elif chains, first match wins
		{"let x = 7\nif x > 10 then\nprint \"high\"\nelif x > 5 then\nprint \"mid\"\nelse\nprint \"low\"\nend\n", "mid\n"},
		{"let x = 20\nif x > 10 then\nprint \"high\"\nelif x > 5 then\nprint \"mid\"\nelse\nprint \"low\"\nend\n", "high\n"},
		{"let x = 1\nif x > 10 then\nprint \"high\"\nelif x > 5 then\nprint \"mid\"\nelse\nprint \"low\"\nend\n", "low\n"},
		// multiple statements per branch
		{"if true then\nlet a = 1\nlet b = 2\nprint a + b\nend\n", "3\n"},
		// condition uses full expression grammar
		{"let name = \"nesh\"\nif name == \"nesh\" and 1 < 2 then\nprint \"ok\"\nend\n", "ok\n"},
		// nested ifs
		{"let x = 15\nif x > 10 then\nif x > 20 then\nprint \"way high\"\nelse\nprint \"just high\"\nend\nend\n", "just high\n"},
		// empty branches are legal
		{"if false then\nelse\nend\nprint \"survived\"\n", "survived\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestIfParseErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"if x > 5\nprint 1\nend\n", `1:9: expected "then", got "\n"`},
		{"if true then\nprint 1\n", `3:1: expected "end", got ""`},
		{"elif true then\nend\n", `1:1: expected statement (let, print, if, or fn), got "elif"`},
	}
	for _, c := range cases {
		script, perr := parser.Parse(c.src)
		if perr == nil {
			t.Errorf("%q: expected parse error, got %v", c.src, script)
			continue
		}
		if perr.Error() != c.want {
			t.Errorf("%q: got %q, want %q", c.src, perr.Error(), c.want)
		}
	}
}

func TestIfRuntimeErrorPosition(t *testing.T) {
	_, err := run(t, "if true then\nprint missing\nend\n")
	if err == nil || err.Error() != "2:7: undefined variable: missing" {
		t.Fatalf("got %v, want position inside the taken branch", err)
	}
}

func TestFunctions(t *testing.T) {
	cases := []struct{ src, want string }{
		// define + call, return value
		{"fn add(a, b)\nreturn a + b\nend\nprint add(2, 3)\n", "5\n"},
		// call inside expressions
		{"fn double(x)\nreturn x * 2\nend\nprint double(4) + 1\n", "9\n"},
		{"fn gt(a, b)\nreturn a > b\nend\nif gt(10, 5) then\nprint \"yes\"\nend\n", "yes\n"},
		// implicit return is false
		{"fn noop()\nend\nprint noop()\n", "false\n"},
		{"fn noop()\nend\nif not noop() then\nprint \"falsy\"\nend\n", "falsy\n"},
		// bare return
		{"fn f(x)\nif x > 0 then\nreturn\nend\nreturn 99\nend\nprint f(1)\nprint f(-1)\n", "false\n99\n"},
		// recursion: factorial and fibonacci
		{"fn fact(n)\nif n <= 1 then\nreturn 1\nend\nreturn n * fact(n - 1)\nend\nprint fact(6)\n", "720\n"},
		{"fn fib(n)\nif n < 2 then\nreturn n\nend\nreturn fib(n - 1) + fib(n - 2)\nend\nprint fib(10)\n", "55\n"},
		// params are locals; reads fall back to globals
		{"let g = 100\nfn f(x)\nreturn x + g\nend\nprint f(1)\n", "101\n"},
		// let inside fn is local — global unchanged
		{"let x = \"global\"\nfn f()\nlet x = \"local\"\nreturn x\nend\nprint f()\nprint x\n", "local\nglobal\n"},
		// param shadows global for reads
		{"let x = 1\nfn f(x)\nreturn x + 10\nend\nprint f(5)\nprint x\n", "15\n1\n"},
		// functions compose
		{"fn inc(n)\nreturn n + 1\nend\nprint inc(inc(inc(0)))\n", "3\n"},
		// strings through functions
		{"fn greet(name)\nreturn \"hi \" + name\nend\nprint greet(\"nesh\")\n", "hi nesh\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestFunctionErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"print nope(1)\n", "1:7: undefined function: nope"},
		{"let x = 5\nprint x(1)\n", `2:7: x is not a function (it is 5)`},
		{"fn f(a)\nreturn a\nend\nprint f()\n", "4:7: f expects 1 argument(s), got 0"},
		{"fn f(a)\nreturn a\nend\nprint f(1, 2)\n", "4:7: f expects 1 argument(s), got 2"},
		{"fn f()\nreturn missing\nend\nprint f()\n", "2:8: undefined variable: missing"},
		{"return 5\n", "1:1: return outside function"},
	}
	for _, c := range cases {
		_, err := run(t, c.src)
		if err == nil {
			t.Errorf("%q: expected error, got none", c.src)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("%q: got %q, want %q", c.src, err.Error(), c.want)
		}
	}
}

// fakeRunner records commands instead of touching the OS.
type fakeRunner struct {
	calls []struct {
		name  string
		args  []string
		stdin string
	}
	code int
}

func (f *fakeRunner) Run(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var b []byte
	if stdin != nil {
		b, _ = io.ReadAll(stdin)
	}
	f.calls = append(f.calls, struct {
		name  string
		args  []string
		stdin string
	}{name, args, string(b)})
	fmt.Fprintln(stdout, "ran:", name)
	return f.code
}

func TestSystemCommands(t *testing.T) {
	out := &fakeOutput{}
	fr := &fakeRunner{code: 0}
	rt := New(out)
	rt.SetRunner(fr)

	runScript := func(src string) *Error {
		script, perr := parser.Parse(src)
		if perr != nil {
			t.Fatalf("parse error for %q: %v", src, perr)
		}
		return rt.Run(script)
	}

	if err := runScript("git status --short\n"); err != nil || out.b.String() != "ran: git\n" {
		t.Fatalf("command failed: %q, %v", out.b.String(), err)
	}
	if fr.calls[0].name != "git" || len(fr.calls[0].args) != 2 || fr.calls[0].args[0] != "status" || fr.calls[0].args[1] != "--short" {
		t.Fatalf("bad call record: %+v", fr.calls[0])
	}

	// run expression captures the exit code
	out.b.Reset()
	if err := runScript("let code = run deploy staging --dry-run\nprint code == 0\n"); err != nil {
		t.Fatal(err)
	}
	if got := out.b.String(); got != "ran: deploy\ntrue\n" {
		t.Fatalf("run expression: got %q", got)
	}
	last := fr.calls[len(fr.calls)-1]
	if last.name != "deploy" || last.args[0] != "staging" || last.args[1] != "--dry-run" {
		t.Fatalf("run args wrong: %+v", last)
	}

	// nonzero exit code flows through
	fr.code = 3
	out.b.Reset()
	if err := runScript("print run failing-cmd\n"); err != nil || out.b.String() != "ran: failing-cmd\n3\n" {
		t.Fatalf("exit code capture: %q, %v", out.b.String(), err)
	}

	// variable shadowing guard
	if err := runScript("let x = 5\nx foo\n"); err == nil ||
		err.Error() != `2:1: x is a variable, not a command — did you mean print x?` {
		t.Fatalf("shadow guard: %v", err)
	}

	// commands work inside functions and loops
	fr.code = 0
	out.b.Reset()
	if err := runScript("fn ping(host)\nrun ping host\nreturn true\nend\nlet i = 0\nwhile i < 2\nping \"a\"\nlet i = i + 1\nend\n"); err != nil {
		t.Fatal(err)
	}
	if n := len(fr.calls); n < 4 {
		t.Fatalf("expected loop+fn to invoke runner repeatedly, got %d calls", n)
	}
}

func TestNoRunnerError(t *testing.T) {
	_, err := run(t, "git status\n")
	if err == nil || err.Error() != "1:1: system commands are not available here" {
		t.Fatalf("got %v", err)
	}
}

func TestBareCallStatement(t *testing.T) {
	out, err := run(t, "fn deploy(env)\nprint \"deploying to\" env\nend\ndeploy(\"prod\")\n")
	if err != nil || out != "deploying to prod\n" {
		t.Fatalf("bare call statement failed: %q, %v", out, err)
	}
	// bare idents now parse as system commands — without a runner they
	// fail at runtime with a clear error instead of a parse error
	if _, rerr := run(t, "x\n"); rerr == nil {
		t.Fatal("bare ident command should fail without a runner")
	}
}

func TestWhileLoop(t *testing.T) {
	cases := []struct{ src, want string }{
		// countdown
		{"let i = 3\nwhile i > 0\nprint i\nlet i = i - 1\nend\n", "3\n2\n1\n"},
		// condition false from the start: body never runs
		{"while false\nprint \"never\"\nend\nprint \"done\"\n", "done\n"},
		// loop with function calls and return inside body
		{"fn bump(n)\nreturn n + 1\nend\nlet i = 0\nwhile i < 5\nlet i = bump(i)\nend\nprint i\n", "5\n"},
	}
	for _, c := range cases {
		out, err := run(t, c.src)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.src, err)
			continue
		}
		if out != c.want {
			t.Errorf("%q: got %q, want %q", c.src, out, c.want)
		}
	}
}

func TestForIn(t *testing.T) {
	out := &fakeOutput{}
	rt := New(out)
	rt.Define("range", func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expects 1 argument, got %d", len(args))
		}
		n, ok := args[0].(Int)
		if !ok {
			return nil, fmt.Errorf("needs an int, got %s", args[0])
		}
		list := make(List, n)
		for i := range list {
			list[i] = Int(i + 1)
		}
		return list, nil
	})

	rt.SetRunner(&fakeRunner{})
	rt.SetEventSink(nil)
	rt.Define("split", func(args []Value) (Value, error) {
		s, ok := args[0].(Str)
		if !ok || len(args) != 2 {
			return nil, fmt.Errorf("bad args")
		}
		sep, _ := args[1].(Str)
		parts := strings.Split(string(s), string(sep))
		list := make(List, len(parts))
		for i, p := range parts {
			list[i] = Str(p)
		}
		return list, nil
	})

	script, perr := parser.Parse("let total = 0\nfor x in range(4)\nlet total = total + x\nend\nprint total\nfor w in split(\"go go go\", \" \")\nprint w\nend\n")
	if perr != nil {
		t.Fatal(perr)
	}
	rt.SetEventSink(nil)
	if err := rt.Run(script); err != nil {
		t.Fatal(err)
	}
	want := "10\ngo\ngo\ngo\n"
	if out.b.String() != want {
		t.Fatalf("got %q, want %q", out.b.String(), want)
	}

	// iterating a non-list is an error
	bad, _ := parser.Parse("for x in 5\nprint x\nend\n")
	if err := rt.Run(bad); err == nil || !strings.Contains(err.Error(), "cannot iterate 5") {
		t.Fatalf("expected iterate error, got %v", err)
	}

	// loop variable binds in the surrounding scope (like `let`)
	scoped, _ := parser.Parse("for y in split(\"a\", \" \")\nprint y\nend\nprint y\n")
	if err := rt.Run(scoped); err != nil {
		t.Fatal(err)
	}
	if out.b.String() != "10\ngo\ngo\ngo\na\na\n" {
		t.Fatalf("loop var scoping wrong: %q", out.b.String())
	}
}

func TestWhitespaceTolerance(t *testing.T) {
	// Indentation is optional style; blocks are closed by keywords.
	// Tabs, multiple spaces, missing spaces before strings, blank lines,
	// and one-line branches must all parse identically.
	src := "let x=1\n\n\tif    x>0   then\n\t\tprint\"spaced out\"\n   elif x<0 then print \"neg\"\n else print \"zero\"\n end\n"
	out, err := run(t, src)
	if err != nil || out != "spaced out\n" {
		t.Fatalf("whitespace handling broken: got %q, %v", out, err)
	}
}

func TestEventStream(t *testing.T) {
	var events []Event
	out := &fakeOutput{}
	rt := New(out)
	rt.SetRunner(&fakeRunner{code: 3})
	rt.SetEventSink(func(e Event) { events = append(events, e) })

	script, perr := parser.Parse("let n = 2\nprint \"n is\" n\nlet code = run deploy\n")
	if perr != nil {
		t.Fatal(perr)
	}
	if err := rt.Run(script); err != nil {
		t.Fatal(err)
	}

	wantTypes := []string{"let", "print", "command", "let"} // let code = run ... emits both
	if len(events) != len(wantTypes) {
		t.Fatalf("got %d events %+v, want %d", len(events), events, len(wantTypes))
	}
	for i, typ := range wantTypes {
		if events[i].Type != typ {
			t.Errorf("event %d: type %q, want %q", i, events[i].Type, typ)
		}
	}
	if events[1].Text != "n is 2" || events[1].Line != 2 {
		t.Errorf("print event wrong: %+v", events[1])
	}
	if events[2].Name != "deploy" || events[2].Code != 3 {
		t.Errorf("command event wrong: %+v", events[2])
	}

	// errors emit an event too
	events = nil
	bad, _ := parser.Parse("print 1 / 0\n")
	rerr := rt.Run(bad)
	if rerr == nil || len(events) != 1 || events[0].Type != "error" {
		t.Fatalf("error event missing: %v %+v", rerr, events)
	}
}

func TestNoSinkMeansNoEvents(t *testing.T) {
	out := &fakeOutput{}
	rt := New(out)
	script, _ := parser.Parse("let x = 1\nprint x\n")
	if err := rt.Run(script); err != nil { // must not panic without a sink
		t.Fatal(err)
	}
}

func TestComparisonErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"print 1 < \"a\"\n", `1:9: cannot compare 1 and a with "<"`},
		{"print \"a\" + (1 == 1)\n", `1:11: cannot join string and true with "+"`},
		{"print -\"x\"\n", `1:7: cannot negate x`},
	}
	for _, c := range cases {
		_, err := run(t, c.src)
		if err == nil {
			t.Errorf("%q: expected error, got none", c.src)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("%q: got %q, want %q", c.src, err.Error(), c.want)
		}
	}
}
