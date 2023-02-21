package encoding

import (
	"io"
)

// BufferReader reads from a byte slice.
type bufferReader struct {
	buffer []byte
	offset int64 // current reading index
}

func newBufferReader(b []byte) *bufferReader {
	return &bufferReader{b, 0}
}

// Len returns the number of bytes of the unread portion of the buffer.
func (r *bufferReader) Len() int {
	if r.offset >= int64(len(r.buffer)) {
		return 0
	}
	return int(int64(len(r.buffer)) - r.offset)
}

// Size returns the length of the buffer
func (r *bufferReader) Size() int { return len(r.buffer) }

// Read implements the io.Reader interface.
func (r *bufferReader) Read(b []byte) (n int, err error) {
	if r.offset >= int64(len(r.buffer)) {
		return 0, io.EOF
	}
	n = copy(b, r.buffer[r.offset:])
	r.offset += int64(n)
	return
}

// Slice returns a sub-slice of the next n bytes of the underlying buffer
// Mutating the offset
func (r *bufferReader) Slice(n uint) ([]byte, error) {
	// Exceeds the buffer length
	if r.offset+int64(n) > int64(len(r.buffer)) {
		return nil, io.EOF
	}

	cur := r.offset
	r.offset += int64(n)
	return r.buffer[cur:r.offset], nil
}
