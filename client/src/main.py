from protocol import Client, Stream
import time


def main():
    s = Client("127.0.0.1") #make it into an address 
    stream: Stream = s.open()
    stream.write(bytearray("hello, world!", "utf-8"))
    print("Receving....")
    res = stream.read()
    print(res)


if __name__ == "__main__":
    main()
