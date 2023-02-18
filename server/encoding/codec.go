package encoding

import (
	"errors"
	"reflect"
)

// Generic Codec interface for encoding and decoding
type Codec interface {
	Encode(*Encoder, reflect.Value) error
	Decode(*Decoder, reflect.Value) error
}

func GetCodec(v interface{}) (Codec, error) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	return getCodec(rv.Type())
}

func getCodec(t reflect.Type) (Codec, error) {
	switch t.Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.String:
		return new(stringCodec), nil
	case reflect.Bool:
		return new(boolCodec), nil
	case reflect.Int32:
		return new(int32Codec), nil
	case reflect.Int64:
		return new(int64Codec), nil
	case reflect.Uint32:
		return new(uint32Codec), nil
	case reflect.Uint64:
		return new(uint64Codec), nil
	case reflect.Float32:
		return new(float32Codec), nil
	case reflect.Float64:
		return new(float64Codec), nil
	case reflect.Struct:
		return newStructCodec(t)
	}
	return nil, errors.New("Unsupported type " + t.String())
}

func newStructCodec(t reflect.Type) (Codec, error) {
	s := new(structCodec)
	err := s.genCodec(t)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type boolCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolCodec) Encode(e *Encoder, rv reflect.Value) error {
	e.writeBool(rv.Bool())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *boolCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeString(rv.String())
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type int32Codec struct{}

// Encode encodes a value into the encoder.
func (c *int32Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeInt32(int32(rv.Int()))
}

// Decode decodes into a reflect value from the decoder.
func (c *int32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type int64Codec struct{}

// Encode encodes a value into the encoder.
func (c *int64Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeInt64(int64(rv.Int()))
}

// Decode decodes into a reflect value from the decoder.
func (c *int64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type uint32Codec struct{}

// Encode encodes a value into the encoder.
func (c *uint32Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeUint32(uint32(rv.Uint()))
}

// Decode decodes into a reflect value from the decoder.
func (c *uint32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type uint64Codec struct{}

// Encode encodes a value into the encoder.
func (c *uint64Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeUint64(uint64(rv.Uint()))
}

// Decode decodes into a reflect value from the decoder.
func (c *uint64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) Encode(e *Encoder, rv reflect.Value) error {
	e.writeFloat32(float32(rv.Float()))
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) Encode(e *Encoder, rv reflect.Value) error {
	e.writeFloat64(rv.Float())
	return nil
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	return nil
}

type (
	structCodec struct {
		fields []*fieldCodec
	}
	fieldCodec struct {
		index int   // The index of the field used in reflect
		codec Codec // The codec to use for this field
	}
)

func (s *structCodec) genCodec(t reflect.Type) error {
	l := t.NumField()
	fields := []int{}
	for i := 0; i < l; i++ {
		fields = append(fields, i)
	}

	s.fields = make([]*fieldCodec, 0, len(fields))

	for _, i := range fields {
		field := t.Field(i)
		codec, err := getCodec(field.Type)
		if err != nil {
			return err
		}

		// Append since unexported fields are skipped
		s.fields = append(s.fields, &fieldCodec{
			index: i,
			codec: codec,
		})
	}
	return nil
}

// Encode encodes a value into the encoder.
func (s *structCodec) Encode(e *Encoder, rv reflect.Value) (err error) {
	for _, i := range s.fields {

		if err = i.codec.Encode(e, rv.Field(i.index)); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *structCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	// for _, i := range c {
	// 	v := rv.Field(i.Index)
	// 	switch {
	// 	case v.Kind() == reflect.Ptr:
	// 		err = i.Codec.DecodeTo(d, v)
	// 	case v.CanSet():
	// 		err = i.Codec.DecodeTo(d, reflect.Indirect(v))
	// 	}

	// 	if err != nil {
	// 		return
	// 	}
	// }
	return nil
}
