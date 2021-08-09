package writer

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)

type Writer struct {
	out *os.File

	mu  *sync.Mutex
	buf bytes.Buffer
	l   int
}

func New() *Writer {
	return &Writer{
		out: os.Stdout,
		mu:  &sync.Mutex{},
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.clear()
	w.l = w.countLines()
	_, err := w.out.Write(w.buf.Bytes())
	w.buf.Reset()

	return err
}

func (w *Writer) countLines() int {
	i := 0
	nl := byte('\n')

	prev := 0

	fd := int(w.out.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil {
		return 0
	}

	for pos, c := range w.buf.Bytes() {
		if c == nl {
			i++
			l := len(w.buf.Bytes()[prev:pos])
			i += int(l / width)
			prev = pos + 1
		}
	}
	return i
}

func (w *Writer) clear() {
	ESC := 27
	clear := fmt.Sprintf("%c[%dA%c[2K", ESC, 1, ESC)
	for i := 0; i < w.l; i++ {
		fmt.Fprint(w.out, clear)
	}
}
