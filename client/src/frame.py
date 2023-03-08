import struct
from enum import Enum


class Flag(Enum):
    # flags for the transmission
    """
    SYN: Synchronization flag
    PSH: pushing data immediately instead of waiting
    ACK: Acknowledgement flag
    NOP: Nothing
    FIN: Finish
    """
    SYN = 0
    PSH = 1
    ACK = 2
    NOP = 3
    FIN = 4


class Frame:
    """Represents a unit of transmission in the protocol"""

    flags = set(item for item in Flag)  # creates a set of all the flags

    def __init__(self, flag: Flag, sid: bytes, rid: int, seqid: int, data=bytearray()):
        if flag not in Frame.flags:
            raise ValueError("Invalid flag")

        self.flag = flag.value.to_bytes(1, byteorder="little")
        self.sid = sid
        self.rid = rid
        self.seqid = seqid
        self.data = data

        self.buffer = bytearray(self.flag)  # Encode flag
        self.buffer.extend(
            len(self.data).to_bytes(2, byteorder="little")
        )  # Encode length of data
        self.buffer.extend(self.rid.to_bytes(4, byteorder="little"))  # Encode rid
        self.buffer.extend(self.sid)
        self.buffer.extend(self.seqid.to_bytes(2, byteorder="little"))  # Encode seqid

        self.buffer.extend(self.data)
        # [flag  len(data) sid  data]


class Header:
    size_of_flag = 1
    size_of_length = 2
    size_of_rid = 4
    size_of_sid = 16
    size_of_seqid = 2
    header_size = (
        size_of_flag + size_of_length + size_of_rid + size_of_sid + size_of_seqid
    )

    def __init__(self, buffer: bytearray):
        self.buf = buffer[: Header.header_size]

    def flag(self):
        # no endian conversion is needed for 1 byte
        return self.buf[0]

    def length(self):
        # 2 bytes
        # < means little-endian H means unsigned short
        return struct.unpack("<H", self.buf[1 : 1 + 2])[0]

    def requestId(self):
        # 4 bytes
        # < means little-endian I means unsigned int32
        return struct.unpack("<I", self.buf[3 : 3 + 4])[0]

    def streamId(self) -> bytes:
        # 4 bytes
        # < means little-endian I means unsigned int32
        return self.buf[7 : 7 + 16]

    def seqId(self):
        # 2 bytes
        # < means little-endian H means unsigned short
        return struct.unpack("<H", self.buf[23:25])[0]
