from protocol import Client, Stream
import time
from message import message, Err


def main():
    s = Client("127.0.0.1") #make it into an address 
    stream: Stream = s.open()
    to_send = message("SearchFlights", {'temp1':'temp2'}, bytearray("hello!", "utf-8"), None)
    to_send = to_send.marhsall()
    print("Sending the marhsalled message: ", to_send.hex())
    stream.write(to_send)
    print("Receving....")
    res = stream.read()
    print(res)


if __name__ == "__main__":
    main()
