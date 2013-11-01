package riff

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func compare(t *testing.T, a, b *Chunk) {
	if a.ID != b.ID {
		t.Errorf("id: %s != %s", a.ID, b.ID)
	}
	if a.Len != b.Len {
		t.Errorf("len: %v != %v", a.Len, b.Len)
	}
	if a.ListID != b.ListID {
		t.Errorf("listId: %s != %s", a.ListID, b.ListID)
	}
	la := len(a.Chunks)
	lb := len(b.Chunks)
	if la != lb {
		t.Errorf("number of chunks: %v != %v", la, lb)
	}
	for i := 0; i < la && i < lb; i++ {
		compare(t, a.Chunks[i], b.Chunks[i])
	}
}

func TestReader(t *testing.T) {
	f, err := os.Open("data/hand.wav")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}
	defer f.Close()

	c, err := ReadChunk(f)
	if err != nil {
		t.Errorf("ReadFrom: %v", err)
	}

	exp := &Chunk{ID: NewID("RIFF"), Len: 7944,
		ListID: NewID("WAVE"),
		Chunks: []*Chunk{
			{ID: NewID("fmt "), Len: 30},
			{ID: NewID("fact"), Len: 4},
			{ID: NewID("data"), Len: 7800},
			{ID: NewID("LIST"), Len: 74,
				ListID: NewID("INFO"),
				Chunks: []*Chunk{
					{ID: NewID("ISFT"), Len: 62},
				},
			},
		},
	}

	compare(t, exp, c)
}

func TestWriter(t *testing.T) {
	f, err := os.Open("data/hand.wav")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}
	defer f.Close()

	c, err := ReadChunk(f)
	if err != nil {
		t.Errorf("ReadFrom: %v", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		t.Errorf("Seek: %v", err)
	}

	buf := new(bytes.Buffer)
	_, err = c.WriteTo(buf)
	if err != nil {
		t.Errorf("WriteTo: ", err)
	}

	fAll, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("ReadAll: %v", err)
	}
	bAll := buf.Bytes()

	for i := range fAll {
		if fAll[i] != bAll[i] {
			t.Fatalf("wrong char, expected %v got %v", fAll[i], bAll[i])
		}
	}

}
