import socket
import time
import select
from frame import Header, Frame, Flag, EOF
import uuid
from threading import Thread, Lock
from queue import Queue, Empty


class Stream:
    def __init__(self, sess, sid, rid, maxFrameSize, deadline):
        self.session = sess  # the session
        self.sid = sid  # session id
        self.rid = rid  # request id
        self.maxFrameSize = maxFrameSize  # max frame size in the session
        self.deadline = deadline
        self.closed = False
        self.buffers = Queue()  # Thread safe queue
        self.mutex = Lock()

        # We assume a single read and dne event per call
        self.dneEvent = Queue()
        self.closeEvent = Queue()
        self.deadlineEvent = Queue()

    def pushBuffer(self, b: bytearray):
        self.buffers.put(item=b, block=True)

    def notifyDNE(self):
        self.dneEvent.put(True, block=True)

    def notifyClose(self):
        self.closeEvent.put(True, block=True)

    def setDeadline(self, deadline: int = 0):
        def runnable():
            start_time = time.time()
            while time.time() < start_time + deadline:
                time.sleep(1 / 1000)  # yield cpu
            self.deadlineEvent.put(True, block=True)

        t = Thread(target=runnable)
        t.daemon = True
        t.start()

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
        res = self.readIndefinitely()
        res.sort(key=lambda x: x[0])
        for _, b in res:
            dataBuffer.extend(b)

        return dataBuffer

    def readWait(self):
        while True:
            if self.deadline:
                try:
                    _ = self.deadlineEvent.get(block=False)
                    raise TimeoutError("Read timeout")
                except Empty:
                    pass

            try:
                t = self.dneEvent.get(block=False)
                if not self.buffers.empty():
                    self.dneEvent.put(item=t, block=True)
                    return True
                return False
            except Empty:
                pass

            try:
                _ = self.closeEvent.get(block=False)
                if not self.buffers.empty():
                    return True
                return False
            except Empty:
                pass
            time.sleep(1 / 1000)  # yield cpu

    def readIndefinitely(self):
        res = []
        if self.deadline != None:
            self.setDeadline(self.deadline)

        while True:
            try:
                buffer = self.buffers.get(block=False)
                header = Header(buffer)
                res.append(
                    (
                        header.seqId(),
                        buffer[
                            header.header_size : header.header_size + header.length()
                        ],
                    )
                )
            except Empty:
                pass
            hasBuffers = self.readWait()
            if not hasBuffers:
                break

        return res

    def close(self):
        self.notifyClose()
        self.closed = True
        self.session.writeFrame(frame=Frame(Flag.FIN, self.sid, self.rid, 0))


class Session:
    def __init__(self, sock: socket.socket, target: tuple):
        self.sock = sock
        self.target = target
        self.streams = {}
        self.requestId = 0
        self.mtu = 1500
        self.mutex = Lock()
        self.read_thread = Thread(target=self.recv)
        self.read_thread.daemon = True
        self.read_thread.start()

    def recv(self):
        """client implementation of recv only.
        Does not handle SYN.
        """

        def loop():
            while True:
                d, addr = self.sock.recvfrom(self.mtu)
                buffer = bytearray(d)
                header = Header(buffer)
                if header.flag() == Flag.PSH.value:
                    if header.length() <= 0:
                        continue
                    self.mutex.acquire()
                    if self.streamKey(header.sid(), header.rid()) in self.streams:
                        stream: Stream = self.streams[
                            self.streamKey(header.sid(), header.rid())
                        ]
                        stream.pushBuffer(buffer)
                    self.mutex.release()
                elif header.flag() == Flag.DNE.value:
                    self.mutex.acquire()
                    if self.streamKey(header.sid(), header.rid()) in self.streams:
                        stream: Stream = self.streams[
                            self.streamKey(header.sid(), header.rid())
                        ]
                        stream.notifyDNE()
                    self.mutex.release()
                elif header.flag() == Flag.FIN.value:
                    self.mutex.acquire()
                    if self.streamKey(header.sid(), header.rid()) in self.streams:
                        stream: Stream = self.streams[
                            self.streamKey(header.sid(), header.rid())
                        ]
                        stream.close()
                    self.mutex.release()

        try:
            loop()
        except KeyboardInterrupt:
            return
        except Exception as e:
            print(e)

    def streamKey(self, sid: bytes, rid: int):
        return str(f"{str(sid)}{rid}")

    def openWithExisting(self, stream: Stream, deadline: int):
        sid = stream.sid
        # calling the writeframe function to send synchronization
        self.writeFrame(frame=Frame(Flag.SYN, bytes(sid), self.requestId, 0))
        stream = Stream(
            self, bytes(sid), self.requestId, self.mtu - Header.header_size, deadline
        )
        self.mutex.acquire()
        self.streams[self.streamKey(sid, self.requestId)] = stream
        self.requestId += 1
        self.mutex.release()
        return stream

    def open(self, deadline: int = None):
        sid = uuid.uuid4().bytes[:16]
        # calling the writeframe function to send synchronization
        self.writeFrame(frame=Frame(Flag.SYN, bytes(sid), self.requestId, 0))
        stream = Stream(
            self, bytes(sid), self.requestId, self.mtu - Header.header_size, deadline
        )
        self.mutex.acquire()
        self.streams[self.streamKey(sid, self.requestId)] = stream
        self.requestId += 1
        self.mutex.release()

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

    def open(self, deadline: int = None):
        return self.session.open(deadline=deadline)

    def openWithExisting(self, stream: Stream, deadline: int = None):
        return self.session.openWithExisting(stream=stream, deadline=deadline)
