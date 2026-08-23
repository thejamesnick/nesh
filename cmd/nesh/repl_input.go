package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// errInterrupt is returned when the user cancels a line with Ctrl-C.
var errInterrupt = errors.New("interrupt")

// lineReader abstracts prompt + input so the REPL works identically on
// raw terminals (editing/history) and piped stdin (plain lines).
type lineReader interface {
	ReadLine(prompt string) (string, error)
	Close()
}

// plainReader reads without editing — used when stdin is piped.
type plainReader struct{ in *os.File }

func newPlainReader(in *os.File) *plainReader { return &plainReader{in: in} }

func (p *plainReader) ReadLine(prompt string) (string, error) {
	buf := make([]byte, 0, 1)
	sawInput := false
	for {
		b := make([]byte, 1)
		n, err := p.in.Read(b)
		if n == 0 || err == io.EOF {
			if !sawInput {
				return "", io.EOF
			}
			return string(buf), nil
		}
		if err != nil {
			return "", err
		}
		if b[0] == '\n' {
			return string(buf), nil
		}
		if b[0] != '\r' {
			buf = append(buf, b[0])
			sawInput = true
		}
	}
}

func (p *plainReader) Close() {}

// rawLineReader provides history and basic editing in raw terminal mode:
// typing, backspace, Ctrl-U, up/down history, Ctrl-C cancel, Ctrl-D exit.
type rawLineReader struct {
	in      *os.File
	out     *os.File
	fd      int
	restore *term.State
	history []string
	histIdx int // position while browsing; == len(history) when fresh
}

func newRawLineReader(in, out *os.File) (*rawLineReader, error) {
	fd := int(in.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return &rawLineReader{in: in, out: out, fd: fd, restore: state}, nil
}

func (r *rawLineReader) Close() {
	term.Restore(r.fd, r.restore)
	fmt.Fprint(r.out, "\r\n")
}

func (r *rawLineReader) ReadLine(prompt string) (string, error) {
	if _, err := fmt.Fprintf(r.out, "%s", prompt); err != nil {
		return "", err
	}

	var buf []byte
	r.histIdx = len(r.history)

	redraw := func(s string) {
		fmt.Fprintf(r.out, "\r\x1b[K%s%s", prompt, s)
		buf = []byte(s)
	}

	b := make([]byte, 1)
	for {
		if _, err := r.in.Read(b); err != nil {
			return "", err
		}
		switch c := b[0]; {
		case c == '\r' || c == '\n': // submit
			fmt.Fprint(r.out, "\r\n")
			line := string(buf)
			if line != "" && (len(r.history) == 0 || r.history[len(r.history)-1] != line) {
				r.history = append(r.history, line)
			}
			return line, nil
		case c == 3: // Ctrl-C: abandon the line
			fmt.Fprint(r.out, "^C\r\n")
			return "", errInterrupt
		case c == 4: // Ctrl-D on an empty line: quit
			if len(buf) == 0 {
				fmt.Fprint(r.out, "\r\n")
				return "", io.EOF
			}
		case c == 127 || c == 8: // backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(r.out, "\b \b")
			}
		case c == 21: // Ctrl-U: clear the line
			redraw("")
		case c == 27: // escape sequence: arrows
			seq := make([]byte, 2)
			if n, _ := r.in.Read(seq); n == 2 && seq[0] == '[' {
				switch seq[1] {
				case 'A': // up
					if r.histIdx > 0 {
						r.histIdx--
						redraw(r.history[r.histIdx])
					}
				case 'B': // down
					if r.histIdx < len(r.history)-1 {
						r.histIdx++
						redraw(r.history[r.histIdx])
					} else {
						r.histIdx = len(r.history)
						redraw("")
					}
				}
			}
		case c >= 32: // printable
			buf = append(buf, c)
			fmt.Fprintf(r.out, "%c", c)
		}
	}
}
