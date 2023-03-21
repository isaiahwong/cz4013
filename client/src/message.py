import struct, json, ast


class Error:
    """
    The Error Class:
    The error is the type of error returned from an RPC
    The body is the error message returned from an RPC
    """

    def __init__(self, error: str = "", body: str = ""):
        self.error = error
        self.body = body

    def printerror(self):
        print(self.error)
        print(self.body)


class Message:
    """
    The Message Class:
    This class provides consistent handling of a request.
    The RPC is a string which holds the RPC method to be invoked by the Server
    The Query is a dictionary type which consists of the query parameters for RPC method
    The Body is a bytearray type which contains the byte stream of the actual information regarding the RPC method
    The Error is an Error object type, and it may contain any errors that occurred when calling an RPC method
    """

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
