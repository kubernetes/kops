// Based on net/http/internal
package proxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func TestChunk(t *testing.T) {
	r := NewChunkedReader(bytes.NewBufferString(
		"7\r\nhello, \r\n17\r\nworld! 0123456789abcdef\r\n0\r\n",
	))

	assertNextChunk(t, r, "hello, ")
	assertNextChunk(t, r, "world! 0123456789abcdef")
	assertNoMoreChunks(t, r)
}

func TestIncompleteReadOfChunk(t *testing.T) {
	r := NewChunkedReader(bytes.NewBufferString(
		"7\r\nhello, \r\n17\r\nworld! 0123456789abcdef\r\n0\r\n",
	))

	// Incomplete read of first chunk
	{
		if !r.Next() {
			t.Fatalf("Expected chunk, but ran out early: %v", r.Err())
		}
		if r.Err() != nil {
			t.Fatalf("Error reading chunk: %q", r.Err())
		}
		// Read just 2 bytes
		buf := make([]byte, 2)
		if _, err := io.ReadFull(r.Chunk(), buf[:2]); err != nil {
			t.Fatalf("Error reading first bytes of chunk: %q", err)
		}
		if buf[0] != 'h' || buf[1] != 'e' {
			t.Fatalf("Unexpected first 2 bytes of chunk: %q", buf)
		}
	}

	assertNextChunk(t, r, "world! 0123456789abcdef")
	assertNoMoreChunks(t, r)
}

func TestMalformedChunks(t *testing.T) {
	r := NewChunkedReader(bytes.NewBufferString(
		"7\r\nhello, GARBAGEBYTES17\r\nworld! 0123456789abcdef\r\n0\r\n",
	))

	assertNextChunk(t, r, "hello, ")
	assertError(t, r, "malformed chunked encoding")
}

type charReader byte

// Read an infinite sequence of some char
func (r *charReader) Read(p []byte) (int, error) {
	b := byte(*r)
	for i := range p {
		p[i] = b
	}
	return len(p), nil
}

func TestLargeChunks(t *testing.T) {
	var expected int64 = 1024 * 1024
	chars := charReader('a')
	r := NewChunkedReader(io.MultiReader(
		strings.NewReader(strconv.FormatInt(expected, 16)+"\r\n"),
		&io.LimitedReader{N: expected, R: &chars},
		strings.NewReader("\r\n0\r\n"),
	))

	if !r.Next() {
		t.Fatalf("Expected chunk, but ran out early: %v", r.Err())
	}
	if r.Err() != nil {
		t.Fatalf("Error reading chunk: %q", r.Err())
	}
	n, err := io.Copy(ioutil.Discard, r.Chunk())
	if n != expected {
		t.Errorf("chunk reader read %q; want %q", n, expected)
	}
	if err != nil {
		t.Fatalf("reading chunk: %v", err)
	}
	assertNoMoreChunks(t, r)
}

func TestInvalidChunkSize(t *testing.T) {
	r := NewChunkedReader(bytes.NewBufferString(
		"foobar\r\nhello, \r\n0\r\n",
	))

	assertError(t, r, "invalid byte in chunk length")
}

func TestChunkSizeLineTooLong(t *testing.T) {
	var (
		maxLineLength = 4096
		chunkSize     string
	)
	for i := 0; i < maxLineLength; i++ {
		chunkSize = chunkSize + "0"
	}
	chunkSize = chunkSize + "7"

	r := NewChunkedReader(bytes.NewBufferString(
		chunkSize + "\r\nhello, \r\n0\r\n",
	))

	assertError(t, r, "header line too long")
}

func TestBytesAfterLastChunkAreIgnored(t *testing.T) {
	r := NewChunkedReader(bytes.NewBufferString(
		"7\r\nhello, \r\n0\r\nGARBAGEBYTES",
	))

	assertNextChunk(t, r, "hello, ")
	assertNoMoreChunks(t, r)
}

func assertNextChunk(t *testing.T, r *ChunkedReader, expected string) {
	if !r.Next() {
		t.Fatalf("Expected chunk, but ran out early: %v", r.Err())
	}
	if r.Err() != nil {
		t.Fatalf("Error reading chunk: %q", r.Err())
	}
	data, err := ioutil.ReadAll(r.Chunk())
	if string(data) != expected {
		t.Errorf("chunk reader read %q; want %q", data, expected)
	}
	if err != nil {
		t.Logf(`data: %q`, data)
		t.Fatalf("reading chunk: %v", err)
	}
}

func assertError(t *testing.T, r *ChunkedReader, e string) {
	if r.Next() {
		t.Errorf("Expected failure when reading chunks, but got one")
	}
	if r.Err() == nil || r.Err().Error() != e {
		t.Errorf("chunk reader errored %q; want %q", r.Err(), e)
	}
	data, err := ioutil.ReadAll(r.Chunk())
	if len(data) != 0 {
		t.Errorf("chunk should have been empty. got %q", data)
	}
	if err != nil {
		t.Logf(`data: %q`, data)
		t.Errorf("reading chunk: %v", err)
	}

	if r.Next() {
		t.Errorf("Expected no more chunks, but found too many")
	}
}

func assertNoMoreChunks(t *testing.T, r *ChunkedReader) {
	if r.Next() {
		t.Errorf("Expected no more chunks, but found too many")
	}
	if r.Err() != nil {
		t.Errorf("Expected no error, but found: %q", r.Err())
	}
}
