from protocol import Client, Stream
import time
from flight import Flight, ReserveFlight
from message import message, ErrorMsg, Message, Error
import codec
import datetime
from misc import futuretime

# from threading import Thread, Event
from multiprocessing import Process, Event

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
    res: Message = codec.unmarshal(b, Message())

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


def monitorUpdates(IP_ADD: str, PORT: int, duration: int):
    c = Client(IP_ADD, PORT)
    stream = c.open(None)

    req = Message(
        rpc="MonitorUpdates",
        query={"timestamp": str(int(futuretime(duration * 60 * 60) * 1000))},
    )

    stream.write(codec.marshal(req))

    while True:
        try:
            b = stream.read()
            res: Message = codec.unmarshal(b, Message())
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

        time.sleep(1)  # yield for ctx switch
    return


def reserveFlight(IP_ADD: str, PORT: int, reserved: list):
    c = Client(IP_ADD, PORT)
    flightid = str(input("\nEnter the plane id you wish to reserve a seat in: "))
    seats = int(input("\nEnter the number of seats you wish to reserve: "))
    stream = c.open(5)

    req = Message(rpc="ReserveFlight", query={"id": flightid, "seats": str(seats)})
    stream.write(codec.marshal(req))
    b = stream.read()
    res: Message = codec.unmarshal(b, Message())
    if res.error:
        res.error.printerror()
    else:
        r: ReserveFlight = codec.unmarshal(res.body, ReserveFlight())
        f = r.flight
        reserved.append(r.id)
        print(
            "Your Flight Details are: ",
            f.id,
            f.source,
            f.destination,
            f.airfare,
            f.seat_availability,
            "Your reservation id is: ",
            r.id,
        )


def cancelFlight(IP_ADD: str, PORT: int, reserved: list):
    c = Client(IP_ADD, PORT)

    if len(reserved) == 0:
        print("You have no reserved flights! ")
        return

    print("Your reserved flights are: ", end=" ")
    count = 1
    for item in reserved:
        print(count, ": ", item, end=" ")
        count += 1
    print("\n ")
    reserveid = int(input("Enter the plane id you wish to cancel a seat in: "))
    stream = c.open(5)

    req = Message(rpc="CancelFlight", query={"id": str(reserved[reserveid - 1])})
    stream.write(codec.marshal(req))
    b = stream.read()
    res: Message = codec.unmarshal(b, Message())
    if res.error:
        res.error.printerror()
    else:
        f: Flight() = codec.unmarshal(res.body, Flight())
        reserved.pop(reserveid - 1)
        print(
            "Flight Details: ",
            f.id,
            f.source,
            f.destination,
            f.airfare,
            f.seat_availability,
        )


def print_menu():
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
    print("1: findFlight")
    print("2: reserveFlight")
    print("3: cancelFlight")
    print("4: start monitoring updates")
    print("5: stop monitoring updates")
    print("6: exit")
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")


def main(IP_ADD: str, PORT: int):
    exit = False
    process_flag = False
    reserved = []
    monitoring = Process(target=monitorUpdates, args=(IP_ADD, PORT, 1))
    while not exit:
        try:
            print_menu()
            option = str(input("Enter your choice: "))
            if option == "1":
                rpc_get_flights(IP_ADD, PORT)
            elif option == "2":
                reserveFlight(IP_ADD, PORT, reserved)
            elif option == "3":
                cancelFlight(IP_ADD, PORT, reserved)
            elif option == "4":
                count = 0
                duration = -1
                while count < 3:
                    user_input = input(
                        "For how long will you want to monitor updates? (in hours and maximum 24 hours)."
                    )
                    duration = int(user_input)
                    if duration > 0 and duration < 24:
                        break
                    count += 1

                if count >= 3:
                    print(
                        "You have exceeded the number of tries to set the duration!. Default being set to one hour!"
                    )
                    duration = 1

                if monitoring.is_alive():
                    print("terminating previous monitoring request...")
                    monitoring.terminate()
                    print("Creating new monitoring request....")
                    monitoring = Process(
                        target=monitorUpdates, args=(IP_ADD, PORT, duration)
                    )
                    monitoring.start()
                elif process_flag:
                    print("Creating new monitoring request....")
                    monitoring = Process(
                        target=monitorUpdates, args=(IP_ADD, PORT, duration)
                    )
                    monitoring.start()
                else:
                    print("Your first monitoring request....")
                    monitoring.start()
                process_flag = False
            elif option == "5":
                monitoring.terminate()
                process_flag = True
            elif option == "6":
                print("Stopping monitoring...")
                print("exiting....")
                exit = True
                monitoring.terminate()
                break
            else:
                print("Invalid Option!")
                continue
        except KeyboardInterrupt():
            break


if __name__ == "__main__":
    # IP_ADD = str(input("Enter the ip address: "))
    # PORT = int(input("Enter the port: "))
    IP_ADD = "127.0.0.1"
    PORT = 8080

    main(IP_ADD, PORT)
