package encoding

import (
	"io"
)

// BufferReader for byte stream. Used in decoder to read bytes from the stream
type BufferReader struct {
	buffer []byte
	offset int64 // current reading index
}

// NewBufferReader creates a new BufferReader.
func NewBufferReader(b []byte) *BufferReader {
	return &BufferReader{b, 0}
}

// Len returns the number of bytes of the unread portion of the buffer.
func (r *BufferReader) Len() int {
	if r.offset >= int64(len(r.buffer)) {
		return 0
	}
	return int(int64(len(r.buffer)) - r.offset)
}

// Size returns the length of the buffer
func (r *BufferReader) Size() int { return len(r.buffer) }

// Read implements the io.Reader interface.
func (r *BufferReader) Read(b []byte) (n int, err error) {
	if r.offset >= int64(len(r.buffer)) {
		return 0, io.EOF
	}
	n = copy(b, r.buffer[r.offset:])
	r.offset += int64(n)
	return
}

// Slice returns a sub-slice of the next n bytes of the underlying buffer
// Mutating the offset
func (r *BufferReader) Slice(n uint) ([]byte, error) {
	// Exceeds the buffer length
	if r.offset+int64(n) > int64(len(r.buffer)) {
		return nil, io.EOF
	}

	cur := r.offset
	r.offset += int64(n)
	return r.buffer[cur:r.offset], nil
}
