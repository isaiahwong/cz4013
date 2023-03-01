package encoding

import (
	"encoding/binary"
	"math"
	"reflect"
)

type Decoder struct {
	reader *bufferReader
}

func Unmarshal(data []byte, v interface{}) error {
	d := newDecoder(data)
	err := d.unmarshal(v)
	if err != nil {
		return err
	}
	return nil
}

func newDecoder(data []byte) *Decoder {
	return &Decoder{
		reader: newBufferReader(data),
	}
}

func (d *Decoder) unmarshal(v interface{}) error {
	var err error
	var c Codec

	rv := reflect.ValueOf(v)
	// slice cannot be passed directly and need to be passed by reference
	// as such, the GetCodec will evaluate it as a pointer where the
	// marshal would evaluate it as a slice I.E. not set the ptr flag.
	// Hence we need to convert a slice pointer to a slice to retrieve
	// The actual codec for slice. We then take the Elem() to make it
	// Addressable
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice {
		c, err = GetCodecWithRV(reflect.Indirect(rv))
		rv = reflect.ValueOf(v).Elem()
	} else {
		c, err = GetCodec(v)
	}
	if err != nil {
		return err
	}

	// rv := reflect.Indirect(reflect.ValueOf(v))
	return c.Decode(d, rv)
}

// readBool deserializes boolean values
func (d *Decoder) readBool() (b bool, err error) {
	if err := binary.Read(d.reader, binary.LittleEndian, &b); err != nil {
		return false, err
	}
	return
}

// readString deserializes data to string.
func (d *Decoder) readString() (string, error) {
	var length uint32
	if err := binary.Read(d.reader, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	strBytes := make([]byte, length)
	if err := binary.Read(d.reader, binary.LittleEndian, &strBytes); err != nil {
		return "", err
	}
	return string(strBytes), nil
}

// readInt
func (d *Decoder) readInt() (n int, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

// writeInt32 writes a 8 bit integer
func (d *Decoder) readUint8() (n uint8, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

// readInt32 deserializes data to int32.
func (d *Decoder) readUint32() (n uint32, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

func (d *Decoder) readUint64() (n uint64, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

func (d *Decoder) readInt32() (n int32, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

func (d *Decoder) readInt64() (n int64, err error) {
	if err = binary.Read(d.reader, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return
}

// readFloat32 deserializes data to float32.
func (d *Decoder) readFloat32() (n float32, err error) {
	var bits uint32
	err = binary.Read(d.reader, binary.LittleEndian, &bits)
	if err != nil {
		return 0, err
	}
	n = math.Float32frombits(bits)
	return
}

// readFloat64 for a float64.
func (d *Decoder) readFloat64() (n float64, err error) {
	var bits uint64
	err = binary.Read(d.reader, binary.LittleEndian, &bits)
	if err != nil {
		return 0, err
	}
	n = math.Float64frombits(bits)
	return
}
