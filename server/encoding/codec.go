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
	return getCodec(rv)
}

func getCodec(t reflect.Value) (Codec, error) {
	switch t.Kind() {
	case reflect.String:
		return new(stringCodec), nil
	case reflect.Bool:
		return new(boolCodec), nil
	case reflect.Int:
		return new(intCodec), nil
	case reflect.Int32:
		return new(int32Codec), nil
	case reflect.Int64:
		return new(int64Codec), nil
	case reflect.Uint8:
		return new(uint8Codec), nil
	case reflect.Uint32:
		return new(uint32Codec), nil
	case reflect.Uint64:
		return new(uint64Codec), nil
	case reflect.Float32:
		return new(float32Codec), nil
	case reflect.Float64:
		return new(float64Codec), nil
	case reflect.Ptr:
		return newPtrCodec(t)
	case reflect.Struct:
		return newStructCodec(t)
	case reflect.Map:
		return newMapCodec(t)
	case reflect.Slice:
		switch t.Type().Elem().Kind() {
		case reflect.Ptr:
			return newSlicePtrCodec(t)
		default:
			return newSliceCodec(t)
		}

	}
	return nil, errors.New("Unsupported type " + t.String())
}

type boolCodec struct{}

// Encode encodes a value into the encoder.
func (c *boolCodec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeBool(rv.Bool())
}

// Decode decodes into a reflect value from the decoder.
func (c *boolCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readBool(); err == nil {
		rv.SetBool(b)
	}
	return
}

type stringCodec struct{}

// Encode encodes a value into the encoder.
func (c *stringCodec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeString(rv.String())
}

// Decode decodes into a reflect value from the decoder.
func (c *stringCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if s, err := d.readString(); err == nil {
		rv.SetString(s)
	}
	return
}

type intCodec struct{}

func (c *intCodec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeInt(int(rv.Int()))
}

// Decode decodes into a reflect value from the decoder.
func (c *intCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readInt(); err == nil {
		rv.SetInt(int64(b))
	}
	return
}

type int32Codec struct{}

// Encode encodes a value into the encoder.
func (c *int32Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeInt32(int32(rv.Int()))
}

// Decode decodes into a reflect value from the decoder.
func (c *int32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readInt32(); err == nil {
		rv.SetInt(int64(b))
	}
	return
}

type int64Codec struct{}

// Encode encodes a value into the encoder.
func (c *int64Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeInt64(int64(rv.Int()))
}

// Decode decodes into a reflect value from the decoder.
func (c *int64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readInt64(); err == nil {
		rv.SetInt(b)
	}
	return
}

// Used for bytes. i.e 1 byte = 8 bits
type uint8Codec struct{}

// Encode encodes a value into the encoder.
func (c *uint8Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeUInt8(uint8(rv.Uint()))
}

// Decode decodes into a reflect value from the decoder.
func (c *uint8Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readUint8(); err == nil {
		rv.SetUint(uint64(b))
	}
	return
}

type uint32Codec struct{}

// Encode encodes a value into the encoder.
func (c *uint32Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeUint32(uint32(rv.Uint()))
}

// Decode decodes into a reflect value from the decoder.
func (c *uint32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readUint32(); err == nil {
		rv.SetUint(uint64(b))
	}
	return
}

type uint64Codec struct{}

// Encode encodes a value into the encoder.
func (c *uint64Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeUint64(uint64(rv.Uint()))
}

// Decode decodes into a reflect value from the decoder.
func (c *uint64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readUint64(); err == nil {
		rv.SetUint(uint64(b))
	}
	return
}

type float32Codec struct{}

// Encode encodes a value into the encoder.
func (c *float32Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeFloat32(float32(rv.Float()))
}

// Decode decodes into a reflect value from the decoder.
func (c *float32Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readFloat32(); err == nil {
		rv.SetFloat(float64(b))
	}
	return
}

type float64Codec struct{}

// Encode encodes a value into the encoder.
func (c *float64Codec) Encode(e *Encoder, rv reflect.Value) error {
	return e.writeFloat64(rv.Float())
}

// Decode decodes into a reflect value from the decoder.
func (c *float64Codec) Decode(d *Decoder, rv reflect.Value) (err error) {
	if b, err := d.readFloat64(); err == nil {
		rv.SetFloat(b)
	}
	return
}

// ============================================================================
// Ptr Codec
// ============================================================================
type ptrCodec struct {
	codec Codec
}

func newPtrCodec(t reflect.Value) (Codec, error) {
	switch t.Kind() {
	case reflect.Ptr:
		break
	default:
		return nil, errors.New("Expected pointer type")
	}

	// Handle nils to prevent cases like infinite recursion
	// During decode, we retrieve the codecs as needed
	if t.IsNil() {
		return new(ptrCodec), nil
	}

	codec, err := getCodec(t.Elem())
	if err != nil {
		return nil, err
	}

	return &ptrCodec{
		codec: codec,
	}, nil
}

// Encode encodes a value into the encoder.
func (p *ptrCodec) Encode(e *Encoder, rv reflect.Value) (err error) {
	// Mark as nil if the pointer is nil.
	if rv.IsNil() {
		e.writeBool(true)
		return
	}

	if p.codec == nil {
		return errors.New("Codec not supplied to ptrCodec")
	}
	// Mark as not nil.
	e.writeBool(false)
	err = p.codec.Encode(e, rv.Elem())
	return
}

// Decode decodes into a reflect value from the decoder.
func (p *ptrCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	isNil, err := d.readBool()
	if err != nil {
		return err
	}

	if isNil {
		return
	}

	if rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	// Get the codec at decode stage because we don't want to
	// go into a recursive loop for nested nil pointers.
	if p.codec == nil {
		codec, err := getCodec(rv.Elem())
		if err != nil {
			return err
		}
		p.codec = codec
	}

	return p.codec.Decode(d, rv.Elem())
}

// ============================================================================
// Map Codec
// ============================================================================

type mapCodec struct {
	key Codec
	val Codec
}

// Encode encodes a value into the encoder.
func (m *mapCodec) Encode(e *Encoder, rv reflect.Value) (err error) {
	e.writeUint64(uint64(rv.Len()))
	for _, key := range rv.MapKeys() {
		value := rv.MapIndex(key)

		if err = m.key.Encode(e, key); err != nil {
			return err
		}
		if err = m.val.Encode(e, value); err != nil {
			return err
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (m *mapCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	var kc, vc Codec
	l, err = d.readUint64()
	if err != nil {
		return
	}

	t := rv.Type()
	rv.Set(reflect.MakeMap(t))
	for i := 0; i < int(l); i++ {

		kv := reflect.Indirect(reflect.New(rv.Type().Key()))
		kc = m.key
		if kc == nil {
			if kc, err = getCodec(kv); err != nil {
				return
			}
		}
		if err = kc.Decode(d, kv); err != nil {
			return
		}
		vv := reflect.Indirect(reflect.New(rv.Type().Elem()))
		vc = m.val
		if vc == nil {
			if vc, err = getCodec(vv); err != nil {
				return
			}
		}
		if err = vc.Decode(d, vv); err != nil {
			return
		}

		rv.SetMapIndex(kv, vv)
	}
	return nil
}

func newMapCodec(t reflect.Value) (Codec, error) {
	if t.Len() == 0 {
		return &mapCodec{}, nil
	}

	k := reflect.Indirect(reflect.New(t.Type().Key()))
	key, err := getCodec(k)
	if err != nil {
		return nil, err
	}

	v := reflect.Indirect(reflect.New(t.Type().Elem()))
	val, err := getCodec(v)
	if err != nil {
		return nil, err
	}

	return &mapCodec{
		key: key,
		val: val,
	}, nil
}

// ============================================================================
// Struct Codec
// ============================================================================
type (
	structCodec struct {
		fields []*fieldCodec
	}
	fieldCodec struct {
		index int   // The index of the field used in reflect
		codec Codec // The codec to use for this field
	}
)

func newStructCodec(t reflect.Value) (Codec, error) {
	s := new(structCodec)
	err := s.genCodec(t)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *structCodec) genCodec(t reflect.Value) error {
	l := t.NumField()
	s.fields = []*fieldCodec{}

	for i := 0; i < l; i++ {
		field := t.Field(i)
		codec, err := getCodec(field)
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
			return err
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (s *structCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	for _, i := range s.fields {
		v := rv.Field(i.index)

		switch {
		case v.Kind() == reflect.Ptr:
			err = i.codec.Decode(d, v)
		case v.CanSet():
			err = i.codec.Decode(d, reflect.Indirect(v))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// Slice Codec
// ============================================================================
type sliceCodec struct {
	codec Codec
}

// newSliceCodec returns a new slice codec.
func newSliceCodec(t reflect.Value) (Codec, error) {
	if t.Len() == 0 {
		return &sliceCodec{}, nil
	}
	v := reflect.Indirect(reflect.New(t.Type().Elem()))
	codec, err := getCodec(v)
	if err != nil {
		return nil, err
	}

	return &sliceCodec{
		codec: codec,
	}, nil
}

// Encode encodes a value into the encoder.
func (s *sliceCodec) Encode(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.writeUint64(uint64(l))
	for i := 0; i < l; i++ {
		v := reflect.Indirect(rv.Index(i).Addr())
		if err = s.codec.Encode(e, v); err != nil {
			return
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (s *sliceCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	if l, err = d.readUint64(); err == nil && l > 0 {
		rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {
			v := reflect.Indirect(rv.Index(i))

			// Get Codec at runtime
			if s.codec == nil {
				codec, err := getCodec(v)
				if err != nil {
					return err
				}
				s.codec = codec
			}

			if err = s.codec.Decode(d, v); err != nil {
				return
			}
		}
	}
	return
}

// ============================================================================
// Slice Ptr Codec
// ============================================================================
type slicePtrCodec struct {
	codec Codec
	type_ reflect.Type
}

func newSlicePtrCodec(t reflect.Value) (Codec, error) {
	// Do not provide codec for empty slice
	if t.Len() == 0 {
		return &slicePtrCodec{
			type_: t.Type().Elem().Elem(),
		}, nil
	}

	codec, err := getCodec(t.Index(0).Elem())
	if err != nil {
		return nil, err
	}

	return &slicePtrCodec{
		codec: codec,
		type_: t.Type().Elem().Elem(),
	}, nil
}

// Encode encodes a value into the encoder.
func (c *slicePtrCodec) Encode(e *Encoder, rv reflect.Value) (err error) {
	l := rv.Len()
	e.writeUint64(uint64(l))

	for i := 0; i < l; i++ {
		v := rv.Index(i)
		e.writeBool(v.IsNil())
		if !v.IsNil() {
			if err = c.codec.Encode(e, reflect.Indirect(v)); err != nil {
				return err
			}
		}
	}
	return
}

// Decode decodes into a reflect value from the decoder.
func (c *slicePtrCodec) Decode(d *Decoder, rv reflect.Value) (err error) {
	var l uint64
	var isNil bool
	if l, err = d.readUint64(); err == nil && l > 0 {

		rv.Set(reflect.MakeSlice(rv.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {

			if isNil, err = d.readBool(); !isNil {
				if err != nil {
					return err
				}

				ptr := rv.Index(i)
				ptr.Set(reflect.New(c.type_))

				if c.codec == nil {
					codec, err := getCodec(ptr.Elem())
					if err != nil {
						return err
					}
					c.codec = codec
				}

				if err = c.codec.Decode(d, reflect.Indirect(ptr)); err != nil {
					return
				}
			}
		}
	}
	return
}
