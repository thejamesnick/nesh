package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nesh/internal/parser"
)

// TestIntegrationScripts runs every .nsh in testdata and checks its exact
// output. These are whole-language scripts: control flow, functions,
// recursion, loops, and (via fakeRunner) system commands.
func TestIntegrationScripts(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"hello.nsh", "hello world\n4\n7\n9\n3\n"},
		{"algorithms.nsh", "5! = 120\nfib(10) = 55\niterative: 120\n"},
		{"health.nsh", "host web-01 healthy, load 0.75\nstatus: down\nnegated: false\n"},
		// stdlib.nsh is exercised by internal/builtin's integration test
	}
	for _, c := range cases {
		src, err := os.ReadFile(filepath.Join("..", "..", "testdata", c.file))
		if err != nil {
			t.Fatalf("reading %s: %v", c.file, err)
		}
		script, perr := parser.Parse(string(src))
		if perr != nil {
			t.Errorf("%s: parse error: %v", c.file, perr)
			continue
		}
		out := &fakeOutput{}
		rt := New(out)
		rt.SetRunner(&fakeRunner{code: 0})
		if rerr := rt.Run(script); rerr != nil {
			t.Errorf("%s: runtime error: %v", c.file, rerr)
			continue
		}
		if got := out.b.String(); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.file, got, c.want)
		}
	}
}

// TestPipelineWithCommands exercises the run idiom end to end.
func TestPipelineWithCommands(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "pipeline.nsh"))
	if err != nil {
		t.Fatal(err)
	}
	script, perr := parser.Parse(string(src))
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	out := &fakeOutput{}
	fr := &fakeRunner{code: 0}
	rt := New(out)
	rt.SetRunner(fr)
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	got := out.b.String()
	for _, want := range []string{
		"deploy exited 0",
		"attempt 1 : pushing artifacts",
		"attempt 2 : pushing artifacts",
		"pipeline done",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot: %q", want, got)
		}
	}
	if len(fr.calls) == 0 || fr.calls[0].name != "deploy-staging" {
		t.Errorf("expected deploy-staging command, got %+v", fr.calls)
	}
}
