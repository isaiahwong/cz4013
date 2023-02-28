from protocol import Client, Stream
import time


def main():
    s = Client()
    stream: Stream = s.open()
    stream.write(bytearray("hello, world!", "utf-8"))
    res = stream.read()
    print(res)


if __name__ == "__main__":
    main()
