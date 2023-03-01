import struct, json, ast


class ErrorMsg:
    def __init__(self, err: str, body: str):
        self.err = err
        self.body = body

    def marhsall(self):
        errbytes = self.err.encode("utf-8")
        bodybytes = self.body.encode("utf-8")
        return (
            struct.pack("<H", len(errbytes))
            + struct.pack("<" + str(len(errbytes)) + "s", errbytes)
            + struct.pack("<H", len(bodybytes))
            + struct.pack("<" + str(len(bodybytes)) + "s", bodybytes)
        )

    def printmessage(self):
        print("The Error is : ", self.err, "\nDetails: ", self.body)

    @classmethod
    def unmarshall(cls, data):
        lenerrbytes = struct.unpack_from("<H", data[:2])[0]
        err = data[2 : lenerrbytes + 2].decode("utf-8")
        lenbodybytes = struct.unpack_from(
            "<H", data[lenerrbytes + 2 : 4 + lenerrbytes]
        )[0]
        body = data[4 + lenerrbytes :].decode("utf-8")
        return cls(err, body)


class message:
    def __init__(self, rpc: str, query: dict, body: bytearray(), error: ErrorMsg):
        self.rpc = rpc
        self.query = query
        self.body = body
        self.error = error

    def printmessage(self):
        print(
            "The Message is: ",
            self.rpc,
            "\n Query is: ",
            self.query,
            "\n Body is:",
            self.body,
            " ",
        )
        self.error.printmessage()

    def marhsall(self):
        rpcbytes = self.rpc.encode("utf-8")
        delimiter = ":"
        querybytes = json.dumps(self.query).encode("utf-8")
        if self.error:
            errbytes = self.error.marhsall()
            return bytearray(
                struct.pack("<H", len(rpcbytes))
                + struct.pack("<" + str(len(rpcbytes)) + "s", rpcbytes)
                + struct.pack("<H", len(querybytes))
                + struct.pack("<" + str(len(querybytes)) + "s", querybytes)
                + struct.pack("<H", len(self.body))
                + self.body
                + struct.pack("<H", len(errbytes))
                + errbytes
            )
        else:
            return bytearray(
                struct.pack("<H", len(rpcbytes))
                + struct.pack("<" + str(len(rpcbytes)) + "s", rpcbytes)
                + struct.pack("<H", len(querybytes))
                + struct.pack("<" + str(len(querybytes)) + "s", querybytes)
                + struct.pack("<H", len(self.body))
                + self.body
            )

    @classmethod
    def unmarshall(cls, data):
        lenrpcbytes = struct.unpack_from("<H", data[:2])[0]
        rpc = data[2 : lenrpcbytes + 2].decode("utf-8")
        lenquery = struct.unpack_from("<H", data[2 + lenrpcbytes : 4 + lenrpcbytes])[0]
        strdict = data[4 + lenrpcbytes : 4 + lenrpcbytes + lenquery]
        query = json.loads(strdict.decode("utf-8"))
        lenbody = struct.unpack_from(
            "<H", data[4 + lenrpcbytes + lenquery : 6 + lenrpcbytes + lenquery]
        )[0]
        body = bytearray(
            data[6 + lenrpcbytes + lenquery : 6 + lenrpcbytes + lenquery + lenbody]
        )
        # check if there is an error in the first place
        if data[6 + lenrpcbytes + lenquery + lenbody :]:
            lenerror = struct.unpack_from(
                "<H",
                data[
                    6
                    + lenrpcbytes
                    + lenquery
                    + lenbody : 8
                    + lenrpcbytes
                    + lenquery
                    + lenbody
                ],
            )[0]
            error = ErrorMsg.unmarshall(data[8 + lenrpcbytes + lenquery + lenbody :])
        else:
            error = None
        return cls(rpc, query, body, error)


class Error:
    def __init__(self, error: str = "", body: str = ""):
        self.error = error
        self.body = body


class Message:
    def __init__(
        self,
        rpc: str = "",
        query: dict = {"": ""},
        body: bytearray = bytearray(),
        error: Error = Error(),
    ):
        self.rpc = rpc
        self.query = query
        self.body = body
        self.error = error
