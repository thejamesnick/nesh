package runtime

import (
	"fmt"
	"testing"

	"nesh/internal/parser"
)

// mapFS serves modules from an in-memory map — no real files touched.
type mapFS struct {
	files map[string]string
	reads int
}

func (f *mapFS) ReadFile(path string) ([]byte, error) {
	f.reads++
	if s, ok := f.files[path]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("not found")
}
func (f *mapFS) WriteFile(path string, data []byte) error { f.files[path] = string(data); return nil }
func (f *mapFS) Exists(path string) bool                  { _, ok := f.files[path]; return ok }

func newModuleRuntime(t *testing.T, files map[string]string) (*Runtime, *fakeOutput, *mapFS) {
	t.Helper()
	out := &fakeOutput{}
	rt := New(out)
	fs := &mapFS{files: files}
	rt.SetFileSystem(fs)
	rt.SetRuntimeFactory(func(child *Runtime) { /* tests register their own builtins */ })
	return rt, out, fs
}

func TestImportNamespaced(t *testing.T) {
	rt, out, _ := newModuleRuntime(t, map[string]string{
		"/app/utils.nsh": "let version = \"2.0\"\nfn greet(n)\nreturn \"hi \" + n\nend\n",
	})
	script, perr := parser.Parse("import \"utils.nsh\" as u\nprint u.version\nprint u.greet(\"x\")\n")
	if perr != nil {
		t.Fatal(perr)
	}
	rt.SetBaseDir("/app")
	if err := rt.Run(script); err != nil {
		t.Fatal(err)
	}
	if out.b.String() != "2.0\nhi x\n" {
		t.Fatalf("got %q", out.b.String())
	}
}

func TestImportMergeAndCache(t *testing.T) {
	rt, out, fs := newModuleRuntime(t, map[string]string{
		"/app/lib.nsh": "let tag = \"L\"\nfn f()\nreturn 1\nend\n",
	})
	script, perr := parser.Parse("import \"lib.nsh\"\nimport \"lib.nsh\"\nprint tag\nprint f()\n")
	if perr != nil {
		t.Fatal(perr)
	}
	rt.SetBaseDir("/app")
	if err := rt.Run(script); err != nil {
		t.Fatal(err)
	}
	if out.b.String() != "L\n1\n" {
		t.Fatalf("got %q", out.b.String())
	}
	if fs.reads != 1 {
		t.Fatalf("module executed/loaded %d times, want 1 (cache broken)", fs.reads)
	}
}

func TestImportCycleDetected(t *testing.T) {
	rt, _, _ := newModuleRuntime(t, map[string]string{
		"/app/a.nsh": "import \"b.nsh\"\nlet a = 1\n",
		"/app/b.nsh": "import \"a.nsh\"\nlet b = 2\n",
	})
	script, perr := parser.Parse("import \"a.nsh\"\n")
	if perr != nil {
		t.Fatal(perr)
	}
	rt.SetBaseDir("/app")
	err := rt.Run(script)
	if err == nil || err.Msg != "circular import of b.nsh" && !containsStr(err.Msg, "circular") {
		t.Fatalf("expected circular import error, got %v", err)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestLexicalScopingAcrossModules(t *testing.T) {
	// A module function must see its own module's state, not the caller's.
	rt, out, _ := newModuleRuntime(t, map[string]string{
		"/app/counter.nsh": "let base = 100\nfn add(n)\nreturn base + n\nend\n",
	})
	script, perr := parser.Parse(
		"import \"counter.nsh\" as c\n" +
			"let base = 999\n" + // caller shadows the name — module fn must NOT see this
			"print c.add(5)\n")
	if perr != nil {
		t.Fatal(perr)
	}
	rt.SetBaseDir("/app")
	if err := rt.Run(script); err != nil {
		t.Fatal(err)
	}
	if out.b.String() != "105\n" {
		t.Fatalf("lexical scoping broken: got %q, want 105", out.b.String())
	}
}
