import socket
import time
import select
from frame import Header, Frame, Flag, EOF
import uuid
from threading import Thread, Lock
from queue import Queue, Empty

"""
The stream class provides an active connection between two entities where frames are exchanged 
continuously during its lifetime.
A stream transmits frames where the flags determine the state of a request/response between entities. 
Reordering of frames happens at the stream level. Once a client establishes a stream with a server, 
both entities can exchange frames without a disjoint connection.
"""

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

    # If a deadline is provided, set the deadline
    def setDeadline(self, deadline: int = 0):
        def runnable():
            start_time = time.time()
            while time.time() < start_time + deadline:
                time.sleep(1 / 1000)  # yield cpu
            self.deadlineEvent.put(True, block=True)

        t = Thread(target=runnable)
        t.daemon = True
        t.start()

    # Write to the server 
    def write(self, data=bytearray()):
        bts = data
        seqid = 0
        # create the frame by creating a Frame Object 
        frame = Frame(Flag.PSH, self.sid, self.rid, seqid, data)  
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

    # Read from the server and calls the readIndefinitely function
    def read(self):
        dataBuffer = bytearray()
        res = self.readIndefinitely()
        res.sort(key=lambda x: x[0])
        for _, b in res:
            dataBuffer.extend(b)

        return dataBuffer

    # Read with a deadline and timeout. If nothing is received within the deadline, yield cpu
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


    # If there is no deadline set, read from the server indefinitely
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

    # Close the stream 
    def close(self):
        self.notifyClose()
        self.closed = True
        self.session.writeFrame(frame=Frame(Flag.FIN, self.sid, self.rid, 0))


"""
The session class maintains the UDP connection and streams a the top level. 
The session maintains a hash-map of streams where it takes a SID, RID pair as the key corresponding to the stream.
the session does the following functions:
1. Creates a single stream 
2. Connects to the remote server where it can send its request
"""
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

    """
    client implementation of recv only.
    Does not handle SYN.
    """
    def recv(self):
        
        self.sock.bind(("0.0.0.0", 0))
        
        # To keep reading from the server 
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

    # Returns the streamKey sid and rid 
    def streamKey(self, sid: bytes, rid: int):
        return str(f"{str(sid)}{rid}")


    # Open a session with the existing stream with a deadline
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

    # Open a new session 
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

    # Write a frame to the server 
    def writeFrame(self, frame: Frame, deadline: int = 0):
        self.sock.sendto(frame.buffer, self.target)



"""
The client class keeps track of the address and port of the server
"""
class Client:
    def __init__(self, addr: str, port: int = 8080):
        self.addr = addr
        self.port = port
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        # to reuse ip address to prevent timeout
        self.sock.setsockopt(
            socket.SOL_SOCKET, socket.SO_REUSEADDR, 1
        )  
        # call the session class
        self.session = Session(
            sock=self.sock, target=(addr, port)
        )  

    # Open a new session 
    def open(self, deadline: int = None):
        return self.session.open(deadline=deadline)

    # Open a session with the existing stream
    def openWithExisting(self, stream: Stream, deadline: int = None):
        return self.session.openWithExisting(stream=stream, deadline=deadline)
