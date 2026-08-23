package runtime

import (
	"testing"

	"nesh/internal/parser"
)

// discardOut swallows output with zero retention.
type discardOut struct{}

func (*discardOut) WriteString(s string) (int, error) { return len(s), nil }

func benchScript(b *testing.B, src string) {
	b.Helper()
	script, perr := parser.Parse(src)
	if perr != nil {
		b.Fatal(perr)
	}
	rt := New(&discardOut{})
	rt.SetRunner(&fakeRunner{})
	rt.Define("range", func(args []Value) (Value, error) {
		n, _ := args[0].(Int)
		list := make(List, n)
		for i := range list {
			list[i] = Int(i + 1)
		}
		return list, nil
	})
	rt.Define("split", func(args []Value) (Value, error) { return List{Str("a")}, nil })
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := rt.Run(script); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArithLoop1000(b *testing.B) {
	benchScript(b, "let i = 0\nlet t = 0\nwhile i < 1000\nlet t = t + i\nlet i = i + 1\nend\n")
}

func BenchmarkFnCalls500(b *testing.B) {
	benchScript(b, "fn add(a, b)\nreturn a + b\nend\nlet i = 0\nwhile i < 500\nadd(i, 1)\nlet i = i + 1\nend\n")
}

func BenchmarkForIn100(b *testing.B) {
	benchScript(b, "let t = 0\nfor x in range(100)\nlet t = t + x\nend\n")
}

func BenchmarkIfChain(b *testing.B) {
	benchScript(b, "let x = 5\nif x > 10 then\nprint \"a\"\nelif x > 3 then\nprint \"b\"\nelse\nprint \"c\"\nend\n")
}
