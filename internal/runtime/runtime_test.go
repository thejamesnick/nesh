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
