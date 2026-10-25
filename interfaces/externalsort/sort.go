//go:build !solution

package externalsort

import (
	"container/heap"
	"io"
	"os"
	"sort"
)

type reader struct {
	r io.Reader
}

func (r *reader) ReadLine() (string, error) {
	line := make([]byte, 0, 4096)
	b := make([]byte, 1)
	n, err := r.r.Read(b)
	for n > 0 && (err == nil || err == io.EOF) && b[0] != '\n' {
		line = append(line, b[0])
		n, err = r.r.Read(b)
	}
	return string(line), err
}

func NewReader(r io.Reader) LineReader {
	return &reader{r: r}
}

type writer struct {
	w io.Writer
}

func (w *writer) Write(l string) error {
	_, err := w.w.Write([]byte(l + "\n"))
	return err
}

func NewWriter(w io.Writer) LineWriter {
	return &writer{w: w}
}

type heapItem struct {
	line   string
	reader LineReader
}

type minHeap []heapItem

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].line < h[j].line }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(heapItem))
}
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func Merge(w LineWriter, readers ...LineReader) error {
	h := &minHeap{}
	heap.Init(h)

	for _, r := range readers {
		line, err := r.ReadLine()
		if err != nil && err != io.EOF {
			return err
		}
		if len(line) > 0 || err != io.EOF {
			heap.Push(h, heapItem{
				line:   line,
				reader: r,
			})
		}
	}

	for h.Len() > 0 {
		item := heap.Pop(h).(heapItem)
		if err := w.Write(item.line); err != nil {
			return err
		}
		newLine, err := item.reader.ReadLine()
		if err != nil && err != io.EOF {
			return err
		}
		if len(newLine) > 0 || err != io.EOF {
			heap.Push(h, heapItem{
				line:   newLine,
				reader: item.reader,
			})
		}
	}
	return nil
}

func Sort(w io.Writer, in ...string) error {
	for _, filename := range in {
		if err := SortLinesInFile(filename); err != nil {
			return err
		}
	}

	readers := make([]LineReader, len(in))
	for i, filename := range in {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		readers[i] = NewReader(file)
	}

	return Merge(NewWriter(w), readers...)
}

func SortLinesInFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	r := NewReader(file)
	lines := make([]string, 0)

	line, err := r.ReadLine()
	for err != io.EOF || len(line) > 0 {
		if err != nil && err != io.EOF {
			return err
		}
		if len(line) > 0 || err != io.EOF {
			lines = append(lines, line)
		}
		line, err = r.ReadLine()
	}
	sort.Strings(lines)

	writeFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer writeFile.Close()

	w := NewWriter(writeFile)
	for _, line := range lines {
		if err := w.Write(line); err != nil {
			return err
		}
	}

	return nil
}
