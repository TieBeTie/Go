//go:build !solution

package otp

import (
	"io"
)

type CipherReader struct {
	r    io.Reader
	prng io.Reader
}

func (c *CipherReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)

	if n > 0 {
		key := make([]byte, n)
		_, _ = io.ReadFull(c.prng, key)

		for i := 0; i < n; i++ {
			p[i] ^= key[i]
		}
	}

	return n, err
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &CipherReader{r: r, prng: prng}
}

type CipherWriter struct {
	w    io.Writer
	prng io.Reader
}

func (c *CipherWriter) Write(p []byte) (n int, err error) {
	encrypted := make([]byte, len(p))
	copy(encrypted, p)

	key := make([]byte, len(p))
	_, _ = io.ReadFull(c.prng, key)

	for i := range encrypted {
		encrypted[i] ^= key[i]
	}

	return c.w.Write(encrypted)
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &CipherWriter{w: w, prng: prng}
}
