// Package builtin is the Nesh standard library as plugins.
//
// Contract (MODULES.md): adding a stdlib function = adding a Define call
// here. Zero changes to runtime. Builtins get OS access only via the
// injected shell interfaces — never directly.
package builtin

import (
	"fmt"
	"math"
	"strings"

	"nesh/internal/runtime"
	"nesh/internal/shell"
)

// Value aliases the runtime value interface for readable signatures.
type Value = runtime.Value

// RegisterAll installs the full standard library into rt.
func RegisterAll(rt *runtime.Runtime, fs shell.FileSystem) {
	registerStrings(rt)
	registerMath(rt)
	registerFiles(rt, fs)
}

// wantArgs validates argument count.
func wantArgs(name string, args []Value, min, max int) error {
	got := len(args)
	if got < min || (max >= 0 && got > max) {
		return fmt.Errorf("expects %s argument(s), got %d", arityRange(min, max), got)
	}
	return nil
}

func arityRange(min, max int) string {
	switch {
	case min == max:
		return fmt.Sprintf("%d", min)
	case max < 0:
		return fmt.Sprintf("at least %d", min)
	default:
		return fmt.Sprintf("%d..%d", min, max)
	}
}

func argString(args []Value, i int) (string, error) {
	s, ok := args[i].(runtime.Str)
	if !ok {
		return "", fmt.Errorf("argument %d must be a string, got %s", i+1, args[i])
	}
	return string(s), nil
}

func registerStrings(rt *runtime.Runtime) {
	rt.Define("len", func(args []Value) (Value, error) {
		if err := wantArgs("len", args, 1, 1); err != nil {
			return nil, err
		}
		switch v := args[0].(type) {
		case runtime.Str:
			return runtime.Int(len([]rune(string(v)))), nil
		case runtime.List:
			return runtime.Int(len(v)), nil
		default:
			return nil, fmt.Errorf("len needs a string or list, got %s", v)
		}
	})
	rt.Define("upper", func(args []Value) (Value, error) {
		s, err := oneString("upper", args)
		if err != nil {
			return nil, err
		}
		return runtime.Str(strings.ToUpper(s)), nil
	})
	rt.Define("lower", func(args []Value) (Value, error) {
		s, err := oneString("lower", args)
		if err != nil {
			return nil, err
		}
		return runtime.Str(strings.ToLower(s)), nil
	})
	rt.Define("split", func(args []Value) (Value, error) {
		if err := wantArgs("split", args, 2, 2); err != nil {
			return nil, err
		}
		s, err1 := argString(args, 0)
		sep, err2 := argString(args, 1)
		if err1 != nil {
			return nil, err1
		}
		if err2 != nil {
			return nil, err2
		}
		parts := strings.Split(s, sep)
		list := make(runtime.List, len(parts))
		for i, p := range parts {
			list[i] = runtime.Str(p)
		}
		return list, nil
	})
	rt.Define("join", func(args []Value) (Value, error) {
		if err := wantArgs("join", args, 2, 2); err != nil {
			return nil, err
		}
		list, ok := args[0].(runtime.List)
		if !ok {
			return nil, fmt.Errorf("join needs a list first, got %s", args[0])
		}
		sep, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		parts := make([]string, len(list))
		for i, v := range list {
			parts[i] = v.String()
		}
		return runtime.Str(strings.Join(parts, sep)), nil
	})
	rt.Define("contains", func(args []Value) (Value, error) {
		if err := wantArgs("contains", args, 2, 2); err != nil {
			return nil, err
		}
		s, err1 := argString(args, 0)
		sub, err2 := argString(args, 1)
		if err1 != nil {
			return nil, err1
		}
		if err2 != nil {
			return nil, err2
		}
		return runtime.Bool(strings.Contains(s, sub)), nil
	})
}

func oneString(name string, args []Value) (string, error) {
	if err := wantArgs(name, args, 1, 1); err != nil {
		return "", err
	}
	return argString(args, 0)
}

func registerMath(rt *runtime.Runtime) {
	toFloat := func(v Value) (float64, bool) {
		switch x := v.(type) {
		case runtime.Int:
			return float64(x), true
		case runtime.Float:
			return float64(x), true
		}
		return 0, false
	}
	rt.Define("abs", func(args []Value) (Value, error) {
		if err := wantArgs("abs", args, 1, 1); err != nil {
			return nil, err
		}
		switch x := args[0].(type) {
		case runtime.Int:
			if x < 0 {
				return -x, nil
			}
			return x, nil
		case runtime.Float:
			return runtime.Float(math.Abs(float64(x))), nil
		default:
			return nil, fmt.Errorf("abs needs a number, got %s", x)
		}
	})
	rt.Define("floor", func(args []Value) (Value, error) {
		f, err := oneNumber("floor", args, toFloat)
		if err != nil {
			return nil, err
		}
		return runtime.Int(int64(math.Floor(f))), nil
	})
	rt.Define("round", func(args []Value) (Value, error) {
		f, err := oneNumber("round", args, toFloat)
		if err != nil {
			return nil, err
		}
		return runtime.Int(int64(math.Round(f))), nil
	})
	rt.Define("min", func(args []Value) (Value, error) {
		return extremes("min", args, toFloat, math.Min)
	})
	rt.Define("max", func(args []Value) (Value, error) {
		return extremes("max", args, toFloat, math.Max)
	})
}

func oneNumber(name string, args []Value, toFloat func(Value) (float64, bool)) (float64, error) {
	if err := wantArgs(name, args, 1, 1); err != nil {
		return 0, err
	}
	f, ok := toFloat(args[0])
	if !ok {
		return 0, fmt.Errorf("%s needs a number, got %s", name, args[0])
	}
	return f, nil
}

func extremes(name string, args []Value, toFloat func(Value) (float64, bool), op func(a, b float64) float64) (Value, error) {
	if err := wantArgs(name, args, 1, -1); err != nil {
		return nil, err
	}
	bestF := math.Inf(1)
	if name == "max" {
		bestF = math.Inf(-1)
	}
	allInt := true
	for _, a := range args {
		f, ok := toFloat(a)
		if !ok {
			return nil, fmt.Errorf("%s needs numbers, got %s", name, a)
		}
		if _, isInt := a.(runtime.Int); !isInt {
			allInt = false
		}
		bestF = op(bestF, f)
	}
	if allInt {
		return runtime.Int(int64(bestF)), nil
	}
	return runtime.Float(bestF), nil
}

func registerFiles(rt *runtime.Runtime, fs shell.FileSystem) {
	rt.Define("read", func(args []Value) (Value, error) {
		path, err := oneString("read", args)
		if err != nil {
			return nil, err
		}
		data, rerr := fs.ReadFile(path)
		if rerr != nil {
			return nil, rerr
		}
		return runtime.Str(data), nil
	})
	rt.Define("write", func(args []Value) (Value, error) {
		if err := wantArgs("write", args, 2, 2); err != nil {
			return nil, err
		}
		path, err1 := argString(args, 0)
		data, err2 := argString(args, 1)
		if err1 != nil {
			return nil, err1
		}
		if err2 != nil {
			return nil, err2
		}
		if werr := fs.WriteFile(path, []byte(data)); werr != nil {
			return nil, werr
		}
		return runtime.Bool(true), nil
	})
	rt.Define("exists", func(args []Value) (Value, error) {
		path, err := oneString("exists", args)
		if err != nil {
			return nil, err
		}
		return runtime.Bool(fs.Exists(path)), nil
	})
}
