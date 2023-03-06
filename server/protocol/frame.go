package protocol

import (
	"encoding/binary"
)

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
	Rid  uint32 // Request id used for repeated requests
	Sid  []byte // Stream Id
	Data []byte
}

func (f Frame) Header() header {
	h := header{}
	// Flag
	h[0] = f.Flag

	// Length
	binary.LittleEndian.PutUint16(h[1:], uint16(len(f.Data)))

	// Request id
	binary.LittleEndian.PutUint32(h[3:], f.Rid)

	// Stream id
	copy(h[7:], f.Sid)
	return h
}

func NewFrame(flag byte, sid []byte, rid uint32) Frame {
	return Frame{Flag: flag, Sid: sid, Rid: rid}
}

const (
	sizeOfFlag   = 1
	sizeOfLength = 2
	sizeOfRid    = 4
	sizeOfSid    = 16
	HeaderSize   = sizeOfFlag + sizeOfSid + sizeOfRid + sizeOfLength
)

type header [HeaderSize]byte

func (h header) Flag() byte {
	return h[0]
}

func (h header) Length() uint16 {
	return binary.LittleEndian.Uint16(h[1:])
}

func (h header) RequestID() uint32 {
	return binary.LittleEndian.Uint32(h[3 : 3+4])
}

func (h header) StreamID() []byte {
	return h[7 : 7+16]
}
