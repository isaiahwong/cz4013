import copy
from abc import ABC, abstractmethod
from frame import EOF
import struct

""" Generic Codec interface for encoding and decoding """

""" An encoder class to encode variables of different types """
class Encoder:
    def __init__(self):
        self.out = bytearray()

    # Encodes a boolean variable 
    # < means little-endian ? means boolean
    def write_bool(self, b: bool):
        self.out.extend(struct.pack("<?", 1 if b else 0))

    # Encodes a string variable
    # < means little-endian s means string
    def write_string(self, s: str):
        self.write_int32(len(s))
        self.out.extend(struct.pack(f"<{len(s)}s", s.encode("utf-8")))

    # Encodes an integer variable
    # < means little-endian i means int
    def write_int(self, i: int):
        self.out.extend(struct.pack("<i", i))


    # Encodes a 32 bit integer variable
    # < means little-endian i means int
    def write_int32(self, i: int):
        self.out.extend(struct.pack("<i", int(i)))


    # Encodes a 64 bit integer variable
    # < means little-endian q means int64
    def write_int64(self, i: int):
        self.out.extend(
            struct.pack("<q", int(i))
        )  # required argument is not an integer?


    # Encodes bytes variable
    def write_bytes(self, b: bytes):
        self.out.extend(b)


    # Encodes a bytearray variable
    def write_bytearray(self, b: bytearray):
        self.write_int64(len(b))
        self.out.extend(b)


    # Encodes an unsigned 32 bit integer
    # < means little-endian I means unsigned int32
    def write_uint32(self, i: int):
        self.out.extend(struct.pack("<I", i))

    # Encodes an unsigned 64 bit integer
    # < means little-endian Q means unsigned int64
    def write_uint64(self, i: int):
        self.out.extend(struct.pack("<Q", i))

    # Encodes a 32 bit float variable
    # < means little-endian f means float32
    def write_float32(self, i: float):
        self.out.extend(struct.pack("<f", i))

    # Encodes a 64 bit float variable 
    # < means little-endian d means float32
    def write_float64(self, i: float):
        self.out.extend(struct.pack("<d", i))


""" reads the incoming buffer """
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

""" A Decoder class to decode variables of different types  """
class Decoder:
    def __init__(self, data: bytearray):
        self.buffer_reader: BufferReader = BufferReader(data)

    # Decoding a boolean variable 
    # < means little-endian ? means boolean
    def read_bool(self) -> bool:
        a = self.buffer_reader.read(1)
        return struct.unpack("<?", a)[0]

    # Decoding bytes 
    def read_bytes(self):
        return self.buffer_reader.read(1)


    # Decoding a string variable 
    # < means little-endian s means string
    def read_string(self) -> str:
        l = self.read_int32()
        b = self.buffer_reader.read(l)
        return struct.unpack(f"<{l}s", b)[0].decode("utf-8")


    # Decoding an integer variable
    # < means little-endian i means int
    def read_int(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<i", b)[0]


    # Decoding a 32 bit integer
    # < means little-endian i means int
    def read_int32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<i", b)[0]


    # Decoding a 64 bit integer
    # < means little-endian q means int64 
    def read_int64(self):
        b = self.buffer_reader.read(8)
        return struct.unpack("<q", b)[0]


    # Decoding a byte array 
    def read_bytearray(self):
        l = self.read_int64()
        return self.buffer_reader.read(l)


    # Decoding an unsigned 32 bit integer
    # < means little-endian I means unsigned int32
    def read_uint32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<I", b)[0]


    # Decoding an unsigned 64 bit integer
    # < means little-endian Q means unsigned int64
    def read_uint64(self):
        b = self.buffer_reader.read(8)
        return struct.unpack("<Q", b)[0]

    # Decoding a 32 bit float 
    # < means little-endian f means float32
    def read_float32(self):
        b = self.buffer_reader.read(4)
        return struct.unpack("<f", b)[0]


    #Decoding a 64 bit float 
    # < means little-endian d means float32
    def read_float64(self):
        self.out.extend(struct.pack("<d", i))
        return struct.unpack("<d", b)[0]


#An abstract class is created for encoding and decoding variables of different types 
class Codec(ABC):
    @abstractmethod
    def encode(self, e: Encoder, v: any):
        pass

    @abstractmethod
    def decode(self, d: Decoder) -> any:
        pass


"""
To marshall the content and return a bytearray 
Gets the object type by calling the getCodec function 
Later, it calls the Encoder function to encode the variable 
"""
def marshal(obj) -> bytearray:
    e = Encoder()
    v = getCodec(obj)
    v.encode(e, obj)
    return e.out

"""
To unmarshall the content and return the decoded value 
Gets the object type by calling the getCodec function
Later, it calls the Decoder function to decode the variable  
"""
def unmarshal(b: bytearray, v: any):
    e = Decoder(b)
    v = getCodec(v)
    return v.decode(e)


""" Dynamically finds the type of the variable during runtime """
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


""" 
Implements the Abstract Codec Class to encode and decode
boolean variables 
"""
class BoolCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_bool(v)

    def decode(self, d: Decoder):
        return d.read_bool()

""" 
Implements the Abstract Codec Class to encode and decode
string variables 
"""
class StringCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_string(v)

    def decode(self, d: Decoder):
        return d.read_string()

""" 
Implements the Abstract Codec Class to encode and decode
Integer variables 
"""
class IntCodec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int(v)

    def decode(self, d: Decoder):
        return d.read_int()

""" 
Implements the Abstract Codec Class to encode and decode
32 bit integer variables 
"""
class Int32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int32(v)

    def decode(self, d: Decoder):
        return d.read_int32()

""" 
Implements the Abstract Codec Class to encode and decode
64 bit integer variables 
"""
class Int64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_int64(v)

    def decode(self, d: Decoder):
        return d.read_int64()

""" 
Implements the Abstract Codec Class to encode and decode
unsigned 32 integer variables 
"""
class UInt32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_uint32(v)

    def decode(self, d: Decoder):
        return d.read_uint32()

""" 
Implements the Abstract Codec Class to encode and decode
unsigned 64 bit integer variables 
"""
class UInt64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_uint64(v)

    def decode(self, d: Decoder):
        return d.read_uint64()

""" 
Implements the Abstract Codec Class to encode and decode
32 bit float variables 
"""
class Float32Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_float32(v)

    def decode(self, d: Decoder):
        return d.read_float32()

""" 
Implements the Abstract Codec Class to encode and decode
64 bit float variables 
"""
class Float64Codec(Codec):
    def encode(self, e: Encoder, v: any):
        return e.write_float64(v)

    def decode(self, d: Decoder):
        return d.read_float64()

""" 
Implements the Abstract Codec Class to encode and decode
dictionary variables 
For this class, it iterates through the key, value pairs 
obtains the type dynamically by calling the getCodec function
It then encodes and decodes by iterating through the key,value pairs 
"""
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

""" 
Implements the Abstract Codec Class to encode and decode
list variables 
For this class, it iterates through the items in the list. 
It finds the type by calling the getCodec function.
It encodes and decodes by iterating through the list. 
"""
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

""" 
Implements the Abstract Codec Class to encode and decode
bytearray variables 
"""
class ByteArrayCodec(Codec):
    def encode(self, e: Encoder, obj: bytearray):
        e.write_bytearray(obj)

    def decode(self, d: Decoder):
        return d.read_bytearray()

""" 
Implements the Abstract Codec Class to encode and decode
Object variables 
It iterates through the key,value in the attributes of the object.
It encodes and decodes iteratively through the attribute items. 
"""
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
