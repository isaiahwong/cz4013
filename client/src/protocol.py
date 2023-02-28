import socket
import time
import select
from frame import Header, Frame, Flag


class Stream:
    def __init__(self, sess, sid, maxFrameSize, deadline=5):
        self.session = sess
        self.sid = sid
        self.maxFrameSize = maxFrameSize
        self.deadline = deadline

    def write(self, data=bytearray()):
        frame = Frame(Flag.PSH, self.sid, data)
        bts = data
        while len(bts) > 0:
            size = len(data)
            if size > self.maxFrameSize:
                size = self.maxFrameSize

            frame.Data = bts[:size]
            bts = bts[size:]
            self.session.writeFrame(frame=Frame(Flag.PSH, self.sid, data))
        self.session.writeFrame(frame=Frame(Flag.ACK, self.sid))

    def read(self):
        res = []
        dataBuffer = bytearray()

        while True:
            ready = select.select([self.session.sock], [], [], self.deadline)
            if not ready[0]:
                raise TimeoutError("Read timed out")

            d, addr = self.session.sock.recvfrom(1024)
            buffer = bytearray(d)
            header = Header(buffer)

            if header.flag() == Flag.PSH.value:
                if header.length() <= 0:
                    continue
                dataBuffer.extend(
                    buffer[header.header_size : header.header_size + header.length()]
                )
            elif header.flag() == Flag.ACK.value:
                res.append(dataBuffer)
                dataBuffer = bytearray()
            elif header.flag() == Flag.FIN.value:
                break
        return res

    def close(self):
        self.session.writeFrame(frame=Frame(Flag.FIN, self.sid))


class Session:
    def __init__(self, sock: socket.socket, target: tuple):
        self.sid = 0
        self.sock = sock
        self.target = target
        self.streams = {}

    def open(self):
        self.sid += 2
        self.writeFrame(frame=Frame(Flag.SYN, self.sid))
        self.streams[self.sid] = Stream(self, self.sid, 1024)
        return self.streams[self.sid]

    def writeFrame(self, frame: Frame, deadline: int = 0):
        self.sock.sendto(frame.buffer, self.target)


class Client:
    def __init__(self, addr: str = "127.0.0.1", port: int = 8080):
        self.addr = addr
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.session = Session(sock=self.sock, target=(addr, port))

    def open(self):
        return self.session.open()
