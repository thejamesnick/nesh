package builtin

import (
	"os"
	"strings"
	"testing"

	"nesh/internal/parser"
	"nesh/internal/runtime"
)

// fakeFS avoids touching the real filesystem.
type fakeFS struct {
	files map[string]string
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	if s, ok := f.files[path]; ok {
		return []byte(s), nil
	}
	return nil, &testErr{"not found: " + path}
}

func (f *fakeFS) WriteFile(path string, data []byte) error {
	f.files[path] = string(data)
	return nil
}

func (f *fakeFS) Exists(path string) bool { _, ok := f.files[path]; return ok }

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }

func newTestRuntime() (*runtime.Runtime, *fakeFS) {
	rt := runtime.New(&discard{})
	fs := &fakeFS{files: map[string]string{"greet.txt": "hello"}}
	RegisterAll(rt, fs)
	return rt, fs
}

type discard struct{}

func (*discard) WriteString(s string) (int, error) { return len(s), nil }

// newRuntime builds a runtime with a fresh global scope.
func newRuntime(out runtime.Output) *runtime.Runtime { return runtime.New(out) }

// capture records output for assertions.
type capture struct{ b strings.Builder }

func (c *capture) WriteString(s string) (int, error) { return c.b.WriteString(s) }
func (c *capture) String() string                    { return c.b.String() }

// realOSFS is the real filesystem, used only by the integration test.
type realOSFS struct{}

func (realOSFS) ReadFile(path string) ([]byte, error)     { return os.ReadFile(path) }
func (realOSFS) WriteFile(path string, data []byte) error { return os.WriteFile(path, data, 0o644) }
func (realOSFS) Exists(path string) bool                  { _, err := os.Stat(path); return err == nil }

func evalExpr(t *testing.T, rt *runtime.Runtime, src string) (string, *runtime.Error) {
	t.Helper()
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse %q: %v", src, perr)
	}
	if err := rt.Run(script); err != nil {
		return "", err
	}
	v, ok := rt.Global("__out")
	if !ok {
		t.Fatalf("no __out after running %q", src)
	}
	return v.String(), nil
}

func TestStringBuiltins(t *testing.T) {
	rt, _ := newTestRuntime()
	cases := []struct{ src, want string }{
		{`let __out = len("hello")`, "5"},
		{`let __out = upper("nesh")`, "NESH"},
		{`let __out = lower("MiXeD")`, "mixed"},
		{`let __out = contains("hello world", "wor")`, "true"},
		{`let __out = contains("hello", "xyz")`, "false"},
		{`let l = split("a,b,c", ",")
let __out = len(l)`, "3"},
		{`let l = split("a.b.c", ".")
let __out = join(l, "-")`, "a-b-c"},
		{`let __out = len(split("", ","))`, "1"}, // Go strings.Split semantics
	}
	for _, c := range cases {
		got, err := evalExpr(t, rt, c.src)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestMathBuiltins(t *testing.T) {
	rt, _ := newTestRuntime()
	cases := []struct{ src, want string }{
		{"let __out = abs(-7)", "7"},
		{"let __out = abs(-7.5)", "7.5"},
		{"let __out = floor(3.9)", "3"},
		{"let __out = round(2.5)", "3"},
		{"let __out = min(3, 1, 2)", "1"},
		{"let __out = max(10, 20, 5)", "20"},
		{"let __out = max(0.5, 0.25)", "0.5"},
	}
	for _, c := range cases {
		got, err := evalExpr(t, rt, c.src)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestFileBuiltins(t *testing.T) {
	rt, fs := newTestRuntime()

	got, err := evalExpr(t, rt, `let __out = read("greet.txt")`)
	if err != nil || got != "hello" {
		t.Fatalf("read: %q, %v", got, err)
	}
	if _, err := evalExpr(t, rt, `let __out = read("missing.txt")`); err == nil ||
		!strings.Contains(err.Error(), "not found") {
		t.Fatalf("read missing should fail: %v", err)
	}

	if _, err := evalExpr(t, rt, `let __out = write("out.txt", "data")`); err != nil {
		t.Fatal(err)
	}
	if fs.files["out.txt"] != "data" {
		t.Fatalf("write did not persist: %q", fs.files["out.txt"])
	}

	got, err = evalExpr(t, rt, `let __out = exists("out.txt")`)
	if err != nil || got != "true" {
		t.Fatalf("exists after write: %q, %v", got, err)
	}
	got, err = evalExpr(t, rt, `let __out = exists("nope.txt")`)
	if err != nil || got != "false" {
		t.Fatalf("exists missing: %q, %v", got, err)
	}
}
