// The package riff provides a reader for RIFF (Resource Interchange File
// Format) files as described in
// http://en.wikipedia.org/wiki/Resource_Interchange_File_Format
package riff

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	riff = NewID("RIFF")
	list = NewID("LIST")
)

// Chunk is a chunk of information according to the RIFF specs.
type Chunk struct {
	ID     ID       // Identifier for this chunk
	Len    uint32   // Lenght of this chunk except this field and ID
	Data   []byte   // The data itself
	ListID ID       // Identifier for this RIFF or LIST chunk
	Chunks []*Chunk // Subchunks
}

func (c *Chunk) String() string {
	if len(c.Chunks) == 0 {
		return fmt.Sprintf("%q[%v]", c.ID, c.Len)
	}
	return fmt.Sprintf("%q[%v] {%q: %v}", c.ID, c.Len, c.ListID, c.Chunks)
}

// ReadFrom reads a Chunk from the given reader.
func ReadChunk(r io.Reader) (*Chunk, error) {
	c := new(Chunk)
	// ID
	if err := c.ID.ReadFrom(r); err != nil {
		return nil, err
	}

	// Len
	err := binary.Read(r, binary.LittleEndian, &c.Len)
	if err != nil {
		return nil, err
	}

	// LIST and RIFF contain subchunks
	if c.ID == riff || c.ID == list {
		if err := c.ListID.ReadFrom(r); err != nil {
			return nil, err
		}

		l := c.Len - 4
		for l > 0 {
			sc, err := ReadChunk(r)
			if err != nil {
				return nil, err
			}
			c.Chunks = append(c.Chunks, sc)
			l = l - 8 - sc.Len
		}

		return c, nil
	}

	// Data
	c.Data = make([]byte, c.Len)
	n, err := r.Read(c.Data)
	if err != nil {
		return nil, err
	}
	if n != int(c.Len) {
		return nil, fmt.Errorf("couldn't read all data, read %v bytes of %v", n, c.Len)
	}

	// Pad
	if c.Len%2 != 0 {
		b := make([]byte, 1)
		r.Read(b)
	}

	return c, nil
}

type writer struct {
	w   io.Writer
	err error
	n   int64
}

func (w *writer) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.w.Write(p)
	w.n, w.err = w.n+int64(n), err
	return n, err
}

// WriteTo writes the content of the chunk into the given writer.
func (c *Chunk) WriteTo(w io.Writer) (int64, error) {
	wr := &writer{w: w}

	wr.Write(c.ID[:])
	binary.Write(wr, binary.LittleEndian, c.Len)

	if c.ID == riff || c.ID == list {
		wr.Write(c.ListID[:])
		for i := 0; wr.err == nil && i < len(c.Chunks); i++ {
			c.Chunks[i].WriteTo(wr)
		}
		return wr.n, wr.err
	}

	wr.Write(c.Data)
	if c.Len%2 != 0 {
		w.Write([]byte{'0'})
	}
	return wr.n, wr.err
}

// ID represents a RIFF identifier
type ID [4]byte

// NewID creates a new ID given a 4 characters string. If the size is wrong it panics.
func NewID(s string) ID {
	if len(s) != 4 {
		panic("ID created with wrong length")
	}
	return [4]byte{s[0], s[1], s[2], s[3]}
}

// String returns the string representation of the ID.
func (id *ID) String() string {
	return fmt.Sprintf("%s", id)
}

// ReadFrom reads an ID from the given reader.
func (id *ID) ReadFrom(r io.Reader) error {
	n, err := r.Read(id[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("couldn't read identifier, read %v bytes", n)
	}
	return nil
}
