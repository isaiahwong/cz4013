import socket
import time
import select
from frame import Header, Frame, Flag
import uuid


class Stream:
    def __init__(self, sess, sid, rid, maxFrameSize, deadline):
        self.session = sess  # the session
        self.sid = sid  # session id
        self.rid = rid  # request id
        self.maxFrameSize = maxFrameSize  # max frame size in the session
        self.deadline = deadline

    def write(self, data=bytearray()):
        bts = data
        seqid = 0
        frame = Frame(Flag.PSH, self.sid, self.rid, seqid, data)  # create the frame
        while len(bts) > 0:
            size = len(bts)
            if size > self.maxFrameSize:
                size = self.maxFrameSize
            frame.seqid = seqid
            frame.Data = bts[:size]
            bts = bts[size:]
            self.session.writeFrame(frame=frame)
            seqid += 1
        self.session.writeFrame(frame=Frame(Flag.DNE, self.sid, self.rid, 0))

    def read(self):
        dataBuffer = bytearray()
        res = self.readIndefinitely() if not self.deadline else self.readWithTimeout()
        res.sort(key=lambda x: x[0])
        for _, b in res:
            dataBuffer.extend(b)

        return dataBuffer

    def readIndefinitely(self):
        res = []
        while True:
            d, addr = self.session.sock.recvfrom(self.session.mtu)
            buffer = bytearray(d)
            header = Header(buffer)
            if header.flag() == Flag.FIN.value or header.flag() == Flag.DNE.value:
                break
            self._read(header, buffer, res)

        return res

    def readWithTimeout(self):
        res = []
        while True:
            ready = select.select([self.session.sock], [], [], self.deadline)
            if not ready[0]:
                raise TimeoutError("Read timeout")

            d, addr = self.session.sock.recvfrom(1024)
            buffer = bytearray(d)
            header = Header(buffer)
            if header.flag() == Flag.FIN.value or header.flag() == Flag.DNE.value:
                break
            self._read(header, buffer, res)

        return res

    def _read(self, header: Header, buffer: bytearray, res: list):
        if header.flag() == Flag.PSH.value and header.length() > 0:
            res.append(
                (
                    header.seqId(),
                    buffer[header.header_size : header.header_size + header.length()],
                )
            )

    def close(self):
        self.session.writeFrame(frame=Frame(Flag.FIN, self.sid, self.rid, 0))


class Session:
    def __init__(self, sock: socket.socket, target: tuple):
        self.sock = sock
        self.target = target
        self.streams = {}
        self.requestId = 0
        self.mtu = 1500

    def open(self, deadline: int):
        sid = uuid.uuid4().bytes[:16]
        # calling the writeframe function to send synchronization
        self.writeFrame(frame=Frame(Flag.SYN, bytes(sid), self.requestId, 0))
        stream = Stream(
            self, bytes(sid), self.requestId, self.mtu - Header.header_size, deadline
        )
        self.streams[f"{str(sid)}{self.requestId}"] = stream
        self.requestId += 1
        return stream

    def writeFrame(self, frame: Frame, deadline: int = 0):
        self.sock.sendto(frame.buffer, self.target)


class Client:
    def __init__(self, addr: str, port: int = 8080):
        self.addr = addr
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.sock.setsockopt(
            socket.SOL_SOCKET, socket.SO_REUSEADDR, 1
        )  # to reuse ip address cuz of the timeout issue but can we do this for udp?
        self.session = Session(
            sock=self.sock, target=(addr, port)
        )  # call the session class

    def open(self, deadline: int):
        return self.session.open(deadline=deadline)
