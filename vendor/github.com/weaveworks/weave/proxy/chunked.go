// Based on net/http/internal
package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
)

var (
	ErrLineTooLong        = errors.New("header line too long")
	ErrInvalidChunkLength = errors.New("invalid byte in chunk length")
)

// Unlike net/http/internal.chunkedReader, this has an interface where we can
// handle individual chunks. The interface is based on database/sql.Rows.
func NewChunkedReader(r io.Reader) *ChunkedReader {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &ChunkedReader{r: br}
}

type ChunkedReader struct {
	r     *bufio.Reader
	chunk *io.LimitedReader
	err   error
	buf   [2]byte
}

// Next prepares the next chunk for reading. It returns true on success, or
// false if there is no next chunk or an error happened while preparing
// it. Err should be consulted to distinguish between the two cases.
//
// Every call to Chunk, even the first one, must be preceded by a call to Next.
//
// Calls to Next will discard any unread bytes in the current Chunk.
func (cr *ChunkedReader) Next() bool {
	if cr.err != nil {
		return false
	}

	// Check the termination of the previous chunk
	if cr.chunk != nil {
		// Make sure the remainder is drained, in case the user of this quit
		// reading early.
		if _, cr.err = io.Copy(ioutil.Discard, cr.chunk); cr.err != nil {
			return false
		}

		// Check the next two bytes after the chunk are \r\n
		if _, cr.err = io.ReadFull(cr.r, cr.buf[:2]); cr.err != nil {
			return false
		}
		if cr.buf[0] != '\r' || cr.buf[1] != '\n' {
			cr.err = errors.New("malformed chunked encoding")
			return false
		}
	} else {
		cr.chunk = &io.LimitedReader{R: cr.r}
	}

	// Setup the next chunk
	if n := cr.beginChunk(); n > 0 {
		cr.chunk.N = int64(n)
	} else if cr.err == nil {
		cr.err = io.EOF
	}
	return cr.err == nil
}

// Chunk returns the io.Reader of the current chunk. On each call, this returns
// the same io.Reader for a given chunk.
func (cr *ChunkedReader) Chunk() io.Reader {
	return cr.chunk
}

// Err returns the error, if any, that was encountered during iteration.
func (cr *ChunkedReader) Err() error {
	if cr.err == io.EOF {
		return nil
	}
	return cr.err
}

func (cr *ChunkedReader) beginChunk() uint64 {
	var (
		line []byte
		n    uint64
	)
	// chunk-size CRLF
	line, cr.err = readLine(cr.r)
	if cr.err != nil {
		return 0
	}
	n, cr.err = strconv.ParseUint(string(line), 16, 64)
	if cr.err != nil {
		cr.err = ErrInvalidChunkLength
	}
	return n
}

// Read a line of bytes (up to \n) from b.
// Give up if the line exceeds the buffer size.
// The returned bytes are a pointer into storage in
// the bufio, so they are only valid until the next bufio read.
func readLine(b *bufio.Reader) (p []byte, err error) {
	if p, err = b.ReadSlice('\n'); err != nil {
		// We always know when EOF is coming.
		// If the caller asked for a line, there should be a line.
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		} else if err == bufio.ErrBufferFull {
			err = ErrLineTooLong
		}
		return nil, err
	}
	return bytes.TrimRight(p, " \t\n\r"), nil
}
