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
		{"elif true then\nend\n", `1:1: expected statement (let, print, or if), got "elif"`},
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
