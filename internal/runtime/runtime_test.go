package runtime

import (
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

func TestBareCallStatement(t *testing.T) {
	out, err := run(t, "fn deploy(env)\nprint \"deploying to\" env\nend\ndeploy(\"prod\")\n")
	if err != nil || out != "deploying to prod\n" {
		t.Fatalf("bare call statement failed: %q, %v", out, err)
	}
	// non-call bare identifiers stay a syntax error
	if _, perr := parser.Parse("x\n"); perr == nil {
		t.Fatal("bare ident should not parse as a statement")
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
