package buf

import (
	"bufio"
	"io"
)

type WriteCloser struct {
	*bufio.Writer
}

func NewWriteCloser(w io.Writer) WriteCloser {
	return WriteCloser{
		bufio.NewWriter(w),
	}
}

func (wc WriteCloser) Close() error {
	return wc.Flush()
}
