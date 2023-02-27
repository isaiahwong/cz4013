import socket


class Frame:
    """Represents a unit of transmission in the protocol"""

    @staticmethod
    def SYN():
        # SYN Start of a message
        return int(0).to_bytes(1, "little")

    @staticmethod
    def PSH():
        # Sending of data
        return int(1).to_bytes(1, "little")

    @staticmethod
    def ACK():
        # Sending ACK that represents end of message
        return int(2).to_bytes(1, "little")

    @staticmethod
    def NOP():
        # No operation
        return int(3).to_bytes(1, "little")  # No operation

    @staticmethod
    def FIN():
        # End of stream connection
        return int(4).to_bytes(1, "little")

    def __init__(self, flag, sid, data=bytearray()):
        self.flag = flag
        self.sid = sid
        self.data = data


class Stream:
    def __init__(self, sess, sid, maxFrameSize):
        self.session = sess
        self.sid = sid
        self.maxFrameSize = maxFrameSize

    def write(self, data=bytearray()):
        # TODO: add timer for deadline
        frame = Frame(Frame.PSH(), self.sid, data)
        bts = data
        while len(bts) > 0:
            size = len(data)
            if size > self.maxFrameSize:
                size = self.maxFrameSize

            frame.Data = bts[:size]
            bts = bts[size:]
            self.session.writeFrame(frame=Frame(Frame.PSH(), self.sid, data))
        self.session.writeFrame(frame=Frame(Frame.ACK(), self.sid))

    def close(self):
        self.session.writeFrame(frame=Frame(Frame.FIN(), self.sid))


class Session:
    def __init__(self, sock: socket.socket, target: tuple):
        self.sid = 0
        self.sock = sock
        self.target = target
        self.streams = {}

    def open(self):
        self.sid += 2
        self.writeFrame(frame=Frame(Frame.SYN(), self.sid))
        self.streams[self.sid] = Stream(self, self.sid, 1024)
        return self.streams[self.sid]

    def writeFrame(self, frame: Frame, deadline: int = 0):
        buffer = bytearray(frame.flag)  # Encode flag
        buffer.extend(
            len(frame.data).to_bytes(2, byteorder="little")
        )  # Encode length of data
        buffer.extend(frame.sid.to_bytes(4, byteorder="little"))  # Encode sid
        buffer.extend(frame.data)  # Encode data

        self.sock.sendto(buffer, self.target)


class Client:
    def __init__(self, addr: str = "127.0.0.1", port: int = 8080):
        self.addr = addr
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.session = Session(sock=self.sock, target=(addr, port))

    def open(self):
        return self.session.open()
