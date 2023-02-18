package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

type Decoder struct {
}

func Unmarshal(data []byte, v interface{}) error {
	return nil
}

// readBool deserializes boolean values
func (d *Decoder) readBool(data []byte) (bool, error) {
	if len(data) != 1 {
		return false, errors.New("unexpected data length")
	}
	return data[0] != 0, nil
}

// readString deserializes data to string.
func (d *Decoder) readString(data []byte) string {
	buf := bytes.NewReader(data)
	var length uint32
	err := binary.Read(buf, binary.LittleEndian, &length)
	if err != nil {
		panic(err)
	}
	strBytes := make([]byte, length)
	err = binary.Read(buf, binary.LittleEndian, &strBytes)
	if err != nil {
		panic(err)
	}
	return string(strBytes)
}

// readInt32 deserializes data to int32.
func (d *Decoder) readUInt32(data []byte) (uint32, error) {
	var n uint32
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return n, nil
}

func (d *Decoder) readUInt64(data []byte) (uint64, error) {
	var n uint64
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return n, nil
}

func (d *Decoder) readInt32(data []byte) (int32, error) {
	var n int32
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return n, nil
}

func (d *Decoder) readInt64(data []byte) (int64, error) {
	var n int64
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return n, nil
}

// readFloat32 deserializes data to float32.
func readFloat32(data []byte) (float32, error) {
	if len(data) != 4 {
		return 0, errors.New("unexpected data length")
	}
	bits := binary.LittleEndian.Uint32(data)
	return math.Float32frombits(bits), nil
}

// readFloat64 for a float64.
func readFloat64(data []byte) (float64, error) {
	if len(data) != 8 {
		return 0, errors.New("unexpected data length")
	}
	bits := binary.LittleEndian.Uint64(data)
	return math.Float64frombits(bits), nil
}
