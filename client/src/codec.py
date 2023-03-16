import copy
from abc import ABC, abstractmethod
from frame import EOF
import struct


class Encoder:
    def __init__(self):
        self.out = bytearray()

    def write_bool(self, b: bool):
        self.out.extend(struct.pack("<?", 1 if b else 0))

    def write_string(self, s: str):
        self.write_int32(len(s))
        self.out.extend(struct.pack(f"<{len(s)}s", s.encode("utf-8")))

    def write_int(self, i: int):
        self.out.extend(struct.pack("<i", i))

    def write_int32(self, i: int):
        self.out.extend(struct.pack("<i", int(i)))

    def write_int64(self, i: int):
        self.out.extend(
            struct.pack("<q", int(i))
        )  # required argument is not an integer?

    def write_bytes(self, b: bytes):
        self.out.extend(b)

    def write_bytearray(self, b: bytearray):
        self.write_int64(len(b))
        self.out.extend(b)

    def write_uint32(self, i: int):
        self.out.extend(struct.pack("<I", i))

    def write_uint64(self, i: int):
        self.out.extend(struct.pack("<Q", i))

    def write_float32(self, i: float):
        self.out.extend(struct.pack("<f", i))

    def write_float64(self, i: float):
        self.out.extend(struct.pack("<d", i))


class BufferReader:
    def __init__(self, data: bytearray):
        self.buffer = data
        self.offset = 0

    def read(self, size: int):
        if size == 0:
            return b""
        if self.offset >= len(self.buffer):
            raise EOF("end of buffer")

        self.offset += size
        return self.buffer[self.offset - size : self.offset]


class Decoder:
    def __init__(self, data: bytearray):
        self.buffer_reader: BufferReader = BufferReader(data)

    def read_bool(self) -> bool:
        a = self.buffer_reader.read(1)
        return struct.unpack("<?", a)[0]

    def read_bytes(self):
        return self.buffer_reader.read(1)

    def read_string(self) -> str:
        l = self.read_int32()
        b = self.buffer_reader.read(l)
        return struct.unpack(f"<{l}s", b)[0].decode("utf-8")

    def read_int(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<i", b)[0]

    def read_int32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<i", b)[0]

    def read_int64(self):
        b = self.buffer_reader.read(8)
        return struct.unpack("<q", b)[0]

    def read_bytearray(self):
        l = self.read_int64()
        return self.buffer_reader.read(l)

    def read_uint32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<I", b)[0]

    def read_uint64(self):
        b = self.buffer_reader.read(8)
        return struct.unpack("<Q", b)[0]

    def read_float32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<f", b)[0]

    def read_float64(self):
        self.out.extend(struct.pack("<d", i))
        return struct.unpack("<d", b)[0]


class Codec(ABC):
    @abstractmethod
    def encode(self, e: Encoder, v: any):
        pass

    @abstractmethod
    def decode(self, d: Decoder) -> any:
        pass


def marshal(obj) -> bytearray:
    e = Encoder()
    v = getCodec(obj)
    v.encode(e, obj)
    return e.out


def unmarshal(b: bytearray, v: any):
    e = Decoder(b)
    v = getCodec(v)
    return v.decode(e)


def getCodec(obj) -> Codec:
    if isinstance(obj, dict):
        return DictCodec(obj)
    elif isinstance(obj, list):
        return ListCodec(obj)
    elif isinstance(obj, bool):
        return BoolCodec()
    elif isinstance(obj, int):
        return IntCodec()
    elif isinstance(obj, float):
        return Float32Codec()
    elif isinstance(obj, str):
        return StringCodec()
    elif isinstance(obj, bytearray):
        return ByteArrayCodec()
    else:
        return ObjectCodec(obj)


class BoolCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_bool(v)

    def decode(self, d: Decoder):
        return d.read_bool()


class StringCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_string(v)

    def decode(self, d: Decoder):
        return d.read_string()


class IntCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int(v)

    def decode(self, d: Decoder):
        return d.read_int()


class Int32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int32(v)

    def decode(self, d: Decoder):
        return d.read_int32()


class Int64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int64(v)

    def decode(self, d: Decoder):
        return d.read_int64()


class UInt32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_uint32(v)

    def decode(self, d: Decoder):
        return d.read_uint32()


class UInt64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_uint64(v)

    def decode(self, d: Decoder):
        return d.read_uint64()


class Float32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_float32(v)

    def decode(self, d: Decoder):
        return d.read_float32()


class Float64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_float64(v)

    def decode(self, d: Decoder):
        return d.read_float64()


class DictCodec(Codec):
    """Only homogeneous key value pairs"""

    def __init__(self, d: dict):
        for key, value in d.items():
            self.key_codec = getCodec(key)
            self.value_codec = getCodec(value)
            break

    def encode(self, e: Encoder, d: dict):
        e.write_int64(len(d))
        for key, value in d.items():
            self.key_codec.encode(e, key)
            self.value_codec.encode(e, value)

    def decode(self, d: Decoder):
        res = {}
        length = d.read_int64()

        for _ in range(length):
            key = self.key_codec.decode(d)
            value = self.value_codec.decode(d)
            res[key] = value

        return res


class ListCodec(Codec):
    """Only homogeneous values. We treat list as"""

    def __init__(self, l: list):
        for v in l:
            self.codec = getCodec(v)
            break

    def encode(self, e: Encoder, l: list):
        e.write_int64(len(l))
        for v in l:
            self.codec.encode(e, v)

    def decode(self, d: Decoder):
        l = []
        length = d.read_int64()

        for _ in range(length):
            a = self.codec.decode(d)
            l.append(a)
        return l


class ByteArrayCodec(Codec):
    def encode(self, e: Encoder, obj: bytearray):
        e.write_bytearray(obj)

    def decode(self, d: Decoder):
        return d.read_bytearray()


class ObjectCodec(Codec):
    def __init__(self, obj: any):
        if obj == None:
            return
        self.class_name = obj.__class__.__name__
        self.attributes = vars(obj)
        self.codecs = {}
        self.obj = obj

        for key, value in self.attributes.items():
            self.codecs[key] = getCodec(value)

    def encode(self, e: Encoder, obj: any):
        if obj == None:
            e.write_bool(True)
            return
        e.write_bool(False)

        for k, _ in self.attributes.items():
            self.codecs[k].encode(e, getattr(obj, k))

    def decode(self, d: Decoder):
        isNone = d.read_bool()
        if isNone:
            return None
        res = copy.copy(self.obj)
        for k, v in self.attributes.items():
            setattr(res, k, self.codecs[k].decode(d))
        return res
