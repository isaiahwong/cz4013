from protocol import Client, Stream
import time
from flight import Flight
from message import message, ErrorMsg, Message, Error
import codec
import datetime
from misc import futuretime


#Done with addition of error checking 
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

    # We add a single Flight for codec to know what to unmarshal
    if (res.error):
        res.error.printerror()
    else:
        flights = codec.unmarshal(res.body, [Flight()])

    for f in flights:
        print(f.id, f.source, f.destination, f.airfare, f.seat_availability)




def monitorUpdates():
    c=Client("127.0.0.1",8080)
    stream=c.open()
    print(futuretime()*1000)
    req = Message(
        rpc="MonitorUpdates",
        query={"timestamp":int(futuretime()*1000),"seats":"10"}
    )

    stream.write(codec.marshal(req))

    b = stream.read()
    print(b)
    res: Message = codec.unmarshal(b[0], Message())
    print(res.body)

    if (res.error):
        err: Error = codec.unmarshal(b[0], Error())
        print(err.error)

    flight = codec.unmarshal(res.body, Flight())

    print("New Updated flight: ", flight)


#Done 
def reserveFlight():
    c=Client("127.0.0.1",8080)
    stream=c.open()
    req = Message(
        rpc="ReserveFlight",
        query={"id":"5653","seats":"1"}
    )
    stream.write(codec.marshal(req))
    b = stream.read()
    res: Message = codec.unmarshal(b[0], Message())
    if (res.error):
        res.error.printerror()
    else:
        f = codec.unmarshal(res.body, Flight())
        print("Flight Details: ", f.id, f.source, f.destination, f.airfare, f.seat_availability)



def main():
    #monitorUpdates()
    time.sleep(4)
    rpc_get_flights()
    time.sleep(4)
    reserveFlight()
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
