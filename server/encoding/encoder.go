package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
)

type Encoder struct {
	out io.Writer
}

// Marshal encodes the value v into a byte slice stream.
func Marshal(v interface{}) ([]byte, error) {
	e := newEncoder()
	err := e.marshal(v)

	if err != nil {
		return nil, err
	}

	// safe cast e.out to Buffer
	if buffer, ok := e.out.(*bytes.Buffer); ok {
		return buffer.Bytes(), nil
	}
	return nil, errors.New("Marshal: Failed to cast e.out to bytes.Buffer")
}

// newEncoder creates a new encoder.
func newEncoder() *Encoder {
	var buffer bytes.Buffer
	buffer.Grow(64)

	return &Encoder{
		out: &buffer,
	}
}

// marshal encodes the value v into the buffer.
func (e *Encoder) marshal(v interface{}) error {
	c, err := GetCodec(v)
	if err != nil {
		return err
	}

	// rv := reflect.Indirect(reflect.ValueOf(v))
	return c.Encode(e, reflect.ValueOf(v))
}

// write writes the contents of p into the buffer.
func (e *Encoder) write(p []byte) error {
	_, err := e.out.Write(p)
	return err
}

// writeBool writes a single boolean value into the buffer
func (e *Encoder) writeBool(v bool) error {
	return binary.Write(e.out, binary.LittleEndian, v)
}

// writeString writes a string prefixed with the int size.
func (e *Encoder) writeString(v string) error {
	// Write the size of the string
	err := e.writeInt32(int32(len(v)))
	if err != nil {
		return err
	}
	return binary.Write(e.out, binary.LittleEndian, []byte(v))
}

// writeInt
func (e *Encoder) writeInt(n int) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeInt32 writes a 32 bit integer
func (e *Encoder) writeInt32(n int32) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeInt64 writes a 64 bit integer
func (e *Encoder) writeInt64(n int64) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeInt32 writes a 8 bit integer
func (e *Encoder) writeUInt8(n uint8) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeUint32 writes a 32 bit integer
func (e *Encoder) writeUint32(n uint32) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeUint64 writes a 64 bit integer
func (e *Encoder) writeUint64(n uint64) error {
	err := binary.Write(e.out, binary.LittleEndian, n)
	if err != nil {
		return err
	}
	return nil
}

// writeFloat32 serializes float32. IEEE 754 standard. Assumes float is a finite number
func (e *Encoder) writeFloat32(f float32) error {
	bits := math.Float32bits(f)
	// We only need 4 bytes for 32
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)
	return binary.Write(e.out, binary.LittleEndian, buf)
}

// writeFloat64 serializes float64 or double. IEEE 754 standard. Assumes float is a finite number
func (e *Encoder) writeFloat64(f float64) error {
	bits := math.Float64bits(f)
	// We only need 4 bytes for 32
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint64(buf, bits)
	return binary.Write(e.out, binary.LittleEndian, buf)
}

func (e *Encoder) writeBytes(b []byte) error {
	return binary.Write(e.out, binary.LittleEndian, b)
}
