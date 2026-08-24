package runtime

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nesh/internal/ast"
	"nesh/internal/parser"
	"nesh/internal/shell"
)

// fakeOpener satisfies shell.Opener with in-memory files.
type fakeOpener struct {
	files map[string]*strings.Builder
}

func newFakeOpener() *fakeOpener {
	return &fakeOpener{files: map[string]*strings.Builder{}}
}

type memReader struct{ *strings.Reader }

func (memReader) Close() error { return nil }

func (f *fakeOpener) OpenRead(path string) (io.ReadCloser, error) {
	b, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("no such file")
	}
	return memReader{strings.NewReader(b.String())}, nil
}

type memWriter struct{ b *strings.Builder }

func (w memWriter) Write(p []byte) (int, error) { return w.b.Write(p) }
func (memWriter) Close() error                  { return nil }

func (f *fakeOpener) OpenWrite(path string, appendMode bool) (io.WriteCloser, error) {
	b, ok := f.files[path]
	if !ok || !appendMode {
		b = &strings.Builder{}
		f.files[path] = b
	}
	return memWriter{b: b}, nil
}

func TestRedirectWriteToFile(t *testing.T) {
	opener := newFakeOpener()
	rt := New(&fakeOutput{})
	rt.SetRunner(&fakeRunner{})
	rt.SetFileSystem(fakeFSWithOpener{opener})

	script, perr := parser.Parse("git status > log.txt\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	got := opener.files["log.txt"]
	if got == nil || !strings.Contains(got.String(), "ran: git") {
		t.Fatalf("stdout not redirected: %q", got)
	}
}

// fakeFSWithOpener adapts a fakeOpener to the FileSystem seam.
type fakeFSWithOpener struct{ *fakeOpener }

func (fakeFSWithOpener) ReadFile(path string) ([]byte, error)         { return nil, fmt.Errorf("not implemented") }
func (fakeFSWithOpener) WriteFile(path string, data []byte) error     { return fmt.Errorf("not implemented") }
func (fakeFSWithOpener) Exists(path string) bool                      { return false }

func TestRedirectAppend(t *testing.T) {
	opener := newFakeOpener()
	opener.files["log.txt"] = &strings.Builder{}
	opener.files["log.txt"].WriteString("old\n")

	rt := New(&fakeOutput{})
	fr := &fakeRunner{}
	rt.SetRunner(fr)
	rt.SetFileSystem(fakeFSWithOpener{opener})

	script, perr := parser.Parse("run git status >> log.txt\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	got := opener.files["log.txt"].String()
	if got != "old\nran: git\n" {
		t.Fatalf("append failed: %q", got)
	}
}

func TestRedirectStdinFromFileAndPassthrough(t *testing.T) {
	opener := newFakeOpener()
	opener.files["in.txt"] = &strings.Builder{}
	opener.files["in.txt"].WriteString("file-stdin\n")

	rt := New(&fakeOutput{})
	fr := &fakeRunner{}
	rt.SetRunner(fr)
	rt.SetFileSystem(fakeFSWithOpener{opener})

	// < feeds the child from a file
	script, perr := parser.Parse("cat < in.txt\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if fr.calls[0].stdin != "file-stdin\n" {
		t.Fatalf("< redirect stdin wrong: %q", fr.calls[0].stdin)
	}

	// without <, the script's own stdin flows through
	rt2 := New(&fakeOutput{})
	fr2 := &fakeRunner{}
	rt2.SetRunner(fr2)
	rt2.SetStdin(strings.NewReader("script-stdin\n"))
	script2, perr := parser.Parse("wc -l\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	if rerr := rt2.Run(script2); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	if fr2.calls[0].stdin != "script-stdin\n" {
		t.Fatalf("stdin passthrough wrong: %q", fr2.calls[0].stdin)
	}
}

func TestRedirectWithoutFilesystem(t *testing.T) {
	rt := New(&fakeOutput{})
	rt.SetRunner(&fakeRunner{}) // no fs set

	script, perr := parser.Parse("git status > out.txt\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	rerr := rt.Run(script)
	if rerr == nil || !strings.Contains(rerr.Msg, "redirection needs filesystem access") {
		t.Fatalf("got %v, want redirection-needs-fs error", rerr)
	}
}

func TestOpenReadError(t *testing.T) {
	opener := newFakeOpener()
	rt := New(&fakeOutput{})
	rt.SetRunner(&fakeRunner{})
	rt.SetFileSystem(fakeFSWithOpener{opener})

	script, perr := parser.Parse("cat < missing.txt\n")
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	rerr := rt.Run(script)
	if rerr == nil || !strings.Contains(rerr.Msg, "cannot read missing.txt") {
		t.Fatalf("got %v, want cannot-read error", rerr)
	}
}

// TestRealRedirectionEndToEnd exercises T4.2's done-when with real OS I/O:
// write via >, read it back via <, append via >>.
func TestRealRedirectionEndToEnd(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "deploy.log")

	build := func(src string) (*Runtime, *ast.Script) {
		script, perr := parser.Parse(src)
		if perr != nil {
			t.Fatalf("parse error for %q: %v", src, perr)
		}
		rt := New(&fakeOutput{})
		rt.SetRunner(shell.RealRunner{})
		rt.SetStdin(os.Stdin)
		fs := shell.RealFS{}
		rt.SetFileSystem(fs)
		return rt, script
	}

	// 1. write command output to a file
	rt, script := build("printf \"line1\\nline2\\n\" > \"" + logPath + "\"\n")
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("write redirect failed: %v", rerr)
	}
	data, err := os.ReadFile(logPath)
	if err != nil || string(data) != "line1\nline2\n" {
		t.Fatalf("file content wrong: %q, %v", data, err)
	}

	// 2. feed that file back as stdin and count lines
	rt, script = build("wc -l < \"" + logPath + "\"\n")
	out := &fakeOutput{}
	rt.out = out
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("read redirect failed: %v", rerr)
	}
	if !strings.Contains(out.b.String(), "2") {
		t.Fatalf("expected wc to count 2 lines, got %q", out.b.String())
	}

	// 3. append more output
	rt, script = build("printf \"line3\\n\" >> \"" + logPath + "\"\n")
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("append redirect failed: %v", rerr)
	}
	data, _ = os.ReadFile(logPath)
	if string(data) != "line1\nline2\nline3\n" {
		t.Fatalf("append content wrong: %q", data)
	}
}
