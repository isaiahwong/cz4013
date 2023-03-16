from protocol import Client, Stream
import time
from flight import Flight, ReserveFlight
from message import message, ErrorMsg, Message, Error
import codec
import datetime
from misc import futuretime
from app import App

# from threading import Thread, Event
from multiprocessing import Process, Event

# Done with addition of error checking
def rpc_get_flights(app: App):
    source = str(input("Enter Origin of Flight: "))
    destination = str(input("Enter Destination of Flight: "))

    try:
        flights = app.find_flights(source, destination)

        for f in flights:
            print(f.id, f.source, f.destination, f.airfare, f.seat_availability)

        reserve = ""
        while reserve not in ["Y", "N"]:
            reserve = input("Do you wish to proceed with reserving the flight? (Y/N)")

        if reserve == "Y":
            rpc_reserve_flight(app)
        else:
            return
    except Exception as e:
        print(e)


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


def rpc_reserve_flight(app: App):
    flightid = input("\nEnter the plane id you wish to reserve a seat in: ")
    seats = input("\nEnter the number of seats you wish to reserve: ")
    try:
        r: ReserveFlight = app.reserve_flight(flightid, seats)
        print(
            "Your Flight Details are: ",
            r.flight.id,
            r.flight.source,
            r.flight.destination,
            r.flight.airfare,
            r.flight.seat_availability,
            "Your reservation id is: ",
            r.id,
        )
    except Exception as e:
        print(e)


def rpc_cancel_flight(app: App):
    if len(app.reservations) == 0:
        print("You have no reservations")
        return

    print("Your reserved flights are:\n")
    idxToReservations = []
    count = 0
    for k, v in app.reservations.items():
        print(f"Reservation [{count}]")
        print("Flight Id: ", v.flight.id)
        print("Source: ", v.flight.source)
        print("Destination: ", v.flight.destination)
        print("Airfare: ", v.flight.airfare)
        print("Seats Reserved: ", v.seats_reserved)
        print("Cancelled: ", v.cancelled)
        idxToReservations.append(v.id)
        count += 1

    idx = -1
    while idx < 0 or idx >= len(idxToReservations):
        idx = int(input("Enter the plane id you wish to cancel a seat in: "))

    try:
        rf = app.cancel_flight(idxToReservations[idx])
        if not rf:
            return
        f = rf.flight
        print("Reservation cancelled")
    except Exception as e:
        print(e)


def print_menu():
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
    print("1: Find Flights")
    print("2: Reserve Flights")
    print("3: Cancel Flight")
    print("4: Start monitoring updates")
    print("5: stop monitoring updates")
    print("6: exit")
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")


def main():
    exit = False
    process_flag = False
    a = App()
    while not exit:
        try:
            print_menu()
            option = str(input("Enter your choice: "))
            if option == "1":
                rpc_get_flights(a)
            elif option == "2":
                rpc_reserve_flight(a)
            elif option == "3":
                rpc_cancel_flight(a)
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
                a.monitor_updates(duration, blocking=True)
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
        except KeyboardInterrupt:
            break


if __name__ == "__main__":
    # IP_ADD = str(input("Enter the ip address: "))
    # PORT = int(input("Enter the port: "))
    main()
