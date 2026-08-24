package builtin

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"nesh/internal/parser"
	"nesh/internal/runtime"
	"nesh/internal/shell"
)

type capOut struct{ b bytes.Buffer }

func (c *capOut) WriteString(s string) (int, error) { return c.b.WriteString(s) }

func runCmd(t *testing.T, src string) string {
	t.Helper()
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error for %q: %v", src, perr)
	}
	out := &capOut{}
	rt := runtime.New(out)
	RegisterAll(rt, shell.RealFS{})
	rt.SetFileSystem(shell.RealFS{})
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	return out.b.String()
}

func TestEchoBuiltin(t *testing.T) {
	if got := runCmd(t, "echo hello world\n"); got != "hello world\n" {
		t.Fatalf("got %q", got)
	}
	if got := runCmd(t, "echo -n no-newline\n"); got != "no-newline" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintfBuiltin(t *testing.T) {
	cases := []struct{ src, want string }{
		{"printf \"a\\nb\\n\"\n", "a\nb\n"},
		{"printf \"%s-%s!\" x y\n", "x-y!"},
		{"printf \"100%%\\n\"\n", "100%\n"},
	}
	for _, c := range cases {
		if got := runCmd(t, c.src); got != c.want {
			t.Errorf("%q: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestPrintfRedirectAppend(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/out.txt"
	runCmd(t, "printf \"one\\n\" > \""+p+"\"\n")
	runCmd(t, "printf \"two\\n\" >> \""+p+"\"\n")
	data, err := os.ReadFile(p)
	if err != nil || strings.TrimSpace(string(data)) != "one\ntwo" {
		t.Fatalf("file: %q, %v", data, err)
	}
}

func TestPrintfBuiltinRuns(t *testing.T) {
	// a bare builtin command needs no CommandRunner (no PATH spawn)
	if got := runCmd(t, "printf \"a\\nb\\nc\\n\"\n"); got != "a\nb\nc\n" {
		t.Fatalf("got %q", got)
	}
}
