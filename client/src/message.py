import struct, json, ast


class Error:
    def __init__(self, error: str = "", body: str = ""):
        self.error = error
        self.body = body

    def printerror(self):
        print(self.error)
        print(self.body)


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
