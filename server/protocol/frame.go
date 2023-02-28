package protocol

import "encoding/binary"

// Custom protocol flags for stream in UDP.
const (
	SYN byte = iota // stream open
	PSH             // data push
	ACK             // data acknowledge end of f
	NOP             // no operation
	FIN             // stream close, EOF
)

// Frame used to encapsulate data in UDP.
type Frame struct {
	Flag byte
	Sid  uint32
	Data []byte
}

func (f Frame) Header() header {
	h := header{}
	h[0] = f.Flag
	binary.LittleEndian.PutUint16(h[1:], uint16(len(f.Data)))
	binary.LittleEndian.PutUint32(h[3:], f.Sid)
	return h
}

func NewFrame(flag byte, sid uint32) Frame {
	return Frame{Flag: flag, Sid: sid}
}

const (
	sizeOfFlag   = 1
	sizeOfLength = 2
	sizeOfSid    = 4
	HeaderSize   = sizeOfFlag + sizeOfSid + sizeOfLength
)

type header [HeaderSize]byte

func (h header) Flag() byte {
	return h[0]
}

func (h header) Length() uint16 {
	return binary.LittleEndian.Uint16(h[1:])
}

func (h header) StreamID() uint32 {
	return binary.LittleEndian.Uint32(h[3:])
}
