package runtime

import (
	"strings"
	"testing"

	"nesh/internal/parser"
)

// splitList registers a `range n` builtin producing [1..n], enough to
// exercise for-in without the stdlib.
func splitList(rt *Runtime) {
	rt.Define("range", func(args []Value) (Value, error) {
		n := args[0].(Int)
		list := make(List, n)
		for i := range list {
			list[i] = Int(i + 1)
		}
		return list, nil
	})
}

func TestBreakWhile(t *testing.T) {
	cases := []struct{ src, want string }{
		{"let i = 0\nwhile true\nlet i = i + 1\nif i == 3 then\nbreak\nend\nend\nprint i\n", "3\n"},
		// break before any output
		{"while true\nbreak\nprint \"never\"\nend\n", ""},
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

func TestContinueWhile(t *testing.T) {
	// skip even steps; odds accumulates one "x" per kept iteration
	out, err := run(t, "let i = 0\nlet kept = \"\"\nwhile i < 6\nlet i = i + 1\nif i == 2 then\ncontinue\nend\nif i == 4 then\ncontinue\nend\nif i == 6 then\ncontinue\nend\nlet kept = kept + \"x\"\nend\nprint kept\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "xxx\n" {
		t.Fatalf("got %q, want \"xxx\\n\"", out)
	}
}

func TestBreakContinueForIn(t *testing.T) {
	cases := []struct{ src, want string }{
		{"let total = 0\nfor x in range(5)\nif x == 3 then\nbreak\nend\nlet total = total + x\nend\nprint total\n", "3\n"},                                   // 1+2+3
		{"let total = 0\nfor x in range(5)\nif x == 2 then\ncontinue\nend\nif x == 4 then\ncontinue\nend\nlet total = total + x\nend\nprint total\n", "9\n"}, // 1+3+5
	}
	for _, c := range cases {
		script, perr := parser.Parse(c.src)
		if perr != nil {
			t.Fatalf("%q: parse error: %v", c.src, perr)
		}
		out := &fakeOutput{}
		rt := New(out)
		splitList(rt)
		if rerr := rt.Run(script); rerr != nil {
			t.Errorf("%q: unexpected error %v", c.src, rerr)
			continue
		}
		if got := out.b.String(); got != c.want {
			t.Errorf("%q: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestNestedLoopsBreakInnermost(t *testing.T) {
	src := "for a in range(3)\nfor b in range(3)\nif b == 2 then\nbreak\nend\nprint a b\nend\nend\n"
	want := "1 1\n2 1\n3 1\n"
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	out := &fakeOutput{}
	rt := New(out)
	splitList(rt)
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if got := out.b.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBreakInsideFunctionIsParseError(t *testing.T) {
	// A fn body parses at zero loop depth, so this is rejected at parse
	// time even though the fn sits textually inside a while block.
	src := "while true\nfn f()\nbreak\nend\nend\n"
	_, perr := parser.Parse(src)
	if perr == nil {
		t.Fatal("expected parse error, got none")
	}
	if !strings.Contains(perr.Msg, "loop") {
		t.Fatalf("got %q, want loop-related error", perr.Msg)
	}
}

func TestBreakReturnInteraction(t *testing.T) {
	// return inside a loop inside a function unwinds both.
	src := "fn find()\nfor x in range(6)\nif x == 3 then\nreturn x\nend\nend\nreturn 0\nend\nprint find()\n"
	script, perr := parser.Parse(src)
	if perr != nil {
		t.Fatalf("parse error: %v", perr)
	}
	out := &fakeOutput{}
	rt := New(out)
	splitList(rt)
	if rerr := rt.Run(script); rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if got := out.b.String(); got != "3\n" {
		t.Fatalf("got %q, want \"3\\n\"", got)
	}
}
