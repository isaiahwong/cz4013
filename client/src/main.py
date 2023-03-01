from protocol import Client, Stream
import time
from flight import Flight
from message import message, ErrorMsg, Message
import codec


def rpc_get_flights():
    c = Client("127.0.0.1", 8080)
    stream = c.open()
    req = Message(
        rpc="FindFlights",
        query={"source": "New York", "destination": "Houston"},
    )
    stream.write(codec.marshal(req))
    b = stream.read()
    res: Message = codec.unmarshal(b[0], Message())
    # Define a type for codec to know what to unmarshal
    print(res.body)

    flights = codec.unmarshal(res.body, [Flight()])

    for f in flights:
        print(f.id, f.source, f.destination, f.airfare, f.seat_availability)


def main():
    rpc_get_flights()
    # s = Client("127.0.0.1")  # make it into an address
    # stream: Stream = s.open()
    # err = ErrorMsg("test", "test1")
    # to_send = message(
    #     "SearchFlights", {"temp1": "temp2"}, bytearray("hello!", "utf-8"), err
    # )
    # to_send.printmessage()
    # to_send = to_send.marshall()
    # print("Sending the marshalled message: ", to_send.hex())

    # unmar = message.unmarshall(to_send)
    # unmar.printmessage()
    # stream.write(to_send)
    # print("Receving....")
    # res = stream.read()
    # print(res)


if __name__ == "__main__":
    main()
