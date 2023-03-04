from protocol import Client, Stream
import time
from flight import Flight, ReserveFlight
from message import message, ErrorMsg, Message, Error
import codec
import datetime
from misc import futuretime
import threading

# Done with addition of error checking
def rpc_get_flights(IP_ADD: str, PORT: int):
    c = Client(IP_ADD, PORT)
    stream = c.open(deadline=5)
    source = str(input("Enter Origin of Flight: "))
    destination = str(input("Enter Destination of Flight: "))
    req = Message(
        rpc="FindFlights",
        query={"source": source, "destination": destination},
    )
    stream.write(codec.marshal(req))
    b = stream.read()
    res: Message = codec.unmarshal(b[0], Message())

    # We add a single Flight for codec to know what to unmarshal
    if res.error:
        res.error.printerror()
    else:
        flights = codec.unmarshal(res.body, [Flight()])

    for f in flights:
        print(f.id, f.source, f.destination, f.airfare, f.seat_availability)

    reserve = ""
    while reserve not in ["Y", "N"]:
        reserve = input("Do you wish to proceed with reserving the flight? (Y/N)")

    if reserve == "Y":
        reserveFlight(IP_ADD, PORT)
    else:
        return


def monitorUpdates():
    c = Client("127.0.0.1", 8080)
    stream = c.open(None)

    req = Message(
        rpc="MonitorUpdates",
        query={"timestamp": str(int(futuretime(1 * 60 * 60) * 1000))},
    )

    stream.write(codec.marshal(req))

    while True:
        try:
            b = stream.read()
            res: Message = codec.unmarshal(b[0], Message())
            if res.error:
                res.error.printerror()
                return

            flight: Flight() = codec.unmarshal(res.body, Flight())
            print(
                "New Updated flight: \n",
                flight.id,
                flight.source,
                flight.destination,
                flight.airfare,
                flight.seat_availability,
            )

        except Exception as e:
            print(e)
            pass

        time.sleep(0.1)  # yield for ctx switch


def reserveFlight(IP_ADD: str, PORT: int):
    c = Client(IP_ADD, PORT)
    flightid = str(input("Enter the plane id you wish to reserve a seat in: "))
    seats = int(input("Enter the number of seats you wish to reserve: "))
    stream = c.open(5)

    req = Message(rpc="ReserveFlight", query={"id": flightid, "seats": str(seats)})
    stream.write(codec.marshal(req))
    b = stream.read()
    print(b)
    res: Message = codec.unmarshal(b[0], Message())
    if res.error:
        res.error.printerror()
    else:
        r: ReserveFlight = codec.unmarshal(res.body, ReserveFlight())
        f = r.flight
        print(
            "Flight Details: ",
            f.id,
            f.source,
            f.destination,
            f.airfare,
            f.seat_availability,
        )


def main(IP_ADD: str, PORT: int):
    threading.Thread(target=monitorUpdates).start()
    while True:
        reserveFlight(IP_ADD, PORT)


if __name__ == "__main__":
    # IP_ADD = str(input("Enter the ip address: "))
    # PORT = int(input("Enter the port: "))
    IP_ADD = "127.0.0.1"
    PORT = 8080
    main(IP_ADD, PORT)
