from protocol import Client, Stream
import time
from message import message, ErrorMsg


def main():
    s = Client("127.0.0.1") #make it into an address 
    stream: Stream = s.open()
    err = ErrorMsg("test", "test1")
    to_send = message("SearchFlights", {'temp1':'temp2'}, bytearray("hello!", "utf-8"), err)
    to_send.printmessage()
    to_send = to_send.marhsall()
    print("Sending the marhsalled message: ", to_send.hex())
    
    unmar = message.unmarshall(to_send)
    unmar.printmessage()
    
    stream.write(to_send)
    print("Receving....")
    res = stream.read()
    print(res)


if __name__ == "__main__":
    main()
