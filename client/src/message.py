import struct 

class Err:
    def __int__(self, err: str, body: str):
        self.err = err
        self.body = body
    
    def marhsall(self):
        errbytes = self.err.encode('utf-8')
        bodybytes = self.body.encode('utf-8') 
        return struct.pack('<I', len(errbytes)) + struct.pack('<' + str(len(errbytes))+'s', errbytes) + struct.pack('<I', len(bodybytes)) + struct.pack('<' + str(len(bodybytes))+'s', bodybytes)


class message:
    def __init__(self, rpc: str, query: dict, body: bytearray(), error: Err):
        self.rpc = rpc
        self.query = query
        self.body = body
        self.error = error

    def marhsall(self):
        rpcbytes = self.rpc.encode('utf-8') 
        delimiter = ":"
        querybytes = ''.join([f'{key}{delimiter}{value}\n' for key, value in self.query.items()]).encode('utf-8')
        if self.error:
            errbytes = self.error.marhsall()
            return struct.pack('<I', len(rpcbytes)) + struct.pack('<' + str(len(rpcbytes))+'s', rpcbytes) + struct.pack('<I', len(querybytes))+ struct.pack('<' + str(len(querybytes))+'s', querybytes)+self.body+errbytes
        else:
            return struct.pack('<I', len(rpcbytes))+struct.pack('<' + str(len(rpcbytes))+'s', rpcbytes)+struct.pack('<I', len(querybytes))+struct.pack('<' + str(len(querybytes))+'s', querybytes)+self.body

    


