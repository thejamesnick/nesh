package builtin

import (
	"os"
	"testing"

	"nesh/internal/parser"
)

// TestStdlibIntegrationScript runs testdata/stdlib.nsh end to end with the
// real standard library registered — the full pipeline, no fakes.
func TestStdlibIntegrationScript(t *testing.T) {
	src, err := os.ReadFile("../../testdata/stdlib.nsh")
	if err != nil {
		t.Fatal(err)
	}
	script, perr := parser.Parse(string(src))
	if perr != nil {
		t.Fatalf("parse: %v", perr)
	}
	out := &capture{}
	rt := newRuntime(out)
	RegisterAll(rt, &realOSFS{})
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("runtime error: %v", rerr)
	}
	want := "3 steps\nstep: DEPLOY\nstep: VERIFY\nstep: RELEASE\ndeploy -> verify -> release\ntrue\n2 3 5\n"
	if out.String() != want {
		t.Fatalf("got:\n%q\nwant:\n%q", out.String(), want)
	}
}
