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
	Flag  byte
	Sid   []byte // Stream Id
	Rid   uint32 // Request id used for repeated requests
	SeqId uint16
	Data  []byte
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
	copy(h[3+4:3+4+16], f.Sid)

	// Sequence id
	binary.LittleEndian.PutUint16(h[23:], f.SeqId)

	return h
}

func NewFrame(flag byte, sid []byte, rid uint32, seqId uint16) Frame {
	return Frame{Flag: flag, Sid: sid, Rid: rid, SeqId: seqId}
}

const (
	sizeOfFlag   = 1
	sizeOfLength = 2
	sizeOfSeqId  = 2
	sizeOfRid    = 4
	sizeOfSid    = 16
	HeaderSize   = sizeOfFlag + sizeOfLength + sizeOfRid + sizeOfSid + sizeOfSeqId
)

type header [HeaderSize]byte

func (h header) Flag() byte {
	return h[0]
}

func (h header) Length() uint16 {
	return binary.LittleEndian.Uint16(h[1:])
}

func (h header) RequestID() uint32 {
	return binary.LittleEndian.Uint32(h[3:7])
}

func (h header) StreamID() []byte {
	return h[7:23]
}

func (h header) SeqId() uint16 {
	return binary.LittleEndian.Uint16(h[23:25])
}
