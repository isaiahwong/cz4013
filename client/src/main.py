from protocol import Client, Stream
import time
from flight import Flight, ReserveFlight
from message import Message, Error
import codec
import datetime
from misc import futuretime
from app import App

# from threading import Process, Event
from multiprocessing import Process, Event

"""
A function that uses the App object to call the FindFlights RPC.
It also provides a Command Line Interface with error checking. 
"""
def find_flights(app: App):
    source = str(input("Enter Origin of Flight: "))
    destination = str(input("Enter Destination of Flight: "))

    try:
        flights = app.find_flights(source, destination)

        print()
        for f in flights:
            print(f.id, f.source, f.destination, f.airfare, f.seat_availability)

        print()
        reserve = input("Reserving flight? (Y/N)").upper()
        if reserve == "Y":
            reserve_flight(app)
        else:
            return
    except Exception as e:
        print(f"Error: {e}\n")

"""
A function that uses the App object to call the FindFlight RPC.
It also provides a Command Line Interface with error checking. 
"""
def find_flight(app: App):
    id = str(input("Enter Flight id: "))

    try:
        f = app.find_flight(id)
        print(f)
    except Exception as e:
        print(f"Error: {e}\n")


"""
A function that uses the App object to call the ReserveFlight RPC.
It also provides a Command Line Interface with error checking. 
"""
def reserve_flight(app: App):
    flightid = input("\nEnter Flight id: ")
    seats = input("\nEnter number of seats: ")
    try:
        r: ReserveFlight = app.reserve_flight(flightid, seats)
        print(f"{r}\n")
    except Exception as e:
        print(f"Error: {e}\n")

"""
A function that uses the App object to call the CancelFlight RPC.
It also provides a Command Line Interface with error checking. 
"""
def cancel_flight(app: App):
    if len(app.reservations) == 0:
        print("You have no reservations")
        return

    print("\nReservations:\n")
    app.print_reservations()
    idxToReservations = app.reservations_idx()
    idx = -1
    while idx < 0 or idx >= len(idxToReservations):
        print("Input the reservation number from 0 to ", len(idxToReservations)-1)
        idx = int(input("Select reservation: "))

    try:
        rf = app.cancel_flight(idxToReservations[idx].id)
        if not rf:
            return
        f = rf.flight
        print("Reservation cancelled")
    except Exception as e:
        print(f"Error: {e}\n")

"""
A function that allows the client to view current reservations.
It also provides a Command Line Interface with error checking. 
"""
def view_reservations(app: App):
    if len(app.reservations) == 0:
        print("\nYou have no reservations\n")
        return
    print()
    app.print_reservations()

"""
A function that uses the App object to call the AddMeals RPC.
It also provides a Command Line Interface with error checking. 
"""
def add_meal(app: App):
    if len(app.reservations) == 0:
        print("You have no reservations")
        return
    try:
        print("Reservations:\n")
        app.print_reservations()
        reservations = app.reservations_idx()

        # Get reservations
        r_idx = -1
        while r_idx < 0 or r_idx >= len(reservations):
            r_idx = int(input("Select reservation: "))

        # Get Meals
        meals = app.get_meals()
        print("Meals")
        for i, m in enumerate(meals):
            print(f"Meal[{i}]\n{m}\n")

        meal_idx = -1
        while meal_idx < 0 or meal_idx >= len(meals):
            print("Please select the meal number from 0 to ", len(meals)-1)
            meal_idx = int(input("Select meal: "))

        # Submit RPC request
        rf = app.add_meals(reservations[r_idx].id, meals[meal_idx].id)
        if not rf:
            return
        print(rf)
        print("\nMeal Added\n")
    except Exception as e:
        print(f"Error: {e}\n")

"""
A function that uses the App object to call the MonitorUpdates RPC.
It also provides a Command Line Interface with error checking. 
"""
def monitor_updates(app: App):
    try:
        duration = int(input("Monitor updates Duration (minutes): "))
        app.monitor_updates(duration, blocking=True)
    except Exception as e:
        print(f"Error: {e}\n")

"""
A function that prints out what the client can do. 
It also provides a Command Line Interface with error checking. 
"""
def print_menu():
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
    print("1: Find Flights")
    print("2: Find Flight")
    print("3: Reserve Flights")
    print("4: Cancel Flight")
    print("5: Add Meals")
    print("6: Start monitoring updates")
    print("7: View reservations")
    print("8: exit")
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")


"""
A function that allows the client to use default IP address and Port OR set their own.
Default IP is 127.0.0.1 (localhost) and default port is 8080. 
If you wish to set your own IP and Port, you can enter IP:Port
e.g. 192.168.8.12:2222 
It also provides a Command Line Interface with error checking. 
"""
def clientFactory():
    """Client factory"""
    d = input("Load default Y/N: ").upper()
    if d != "N":
        return App()

    addr = input("Enter remote address: ").split(":")
    remote = addr[0]
    print(remote)
    port = int(addr[1])
    print(port)
    return App(remote=remote, port=port)

""" The Main function that runs a while loop such that the client can choose any of the following options:
1. FindFlights RPC
2. FindFight RPC
3. ReserveFlight RPC
4. CancelFlight RPC
5. AddMeal RPC
6. MonitorUpdates RPC
7. View current reservations
8. exit 
"""
def main():
    a = clientFactory()
    while True:
        try:
            print_menu()
            option = str(input("Enter your choice: "))
            if option == "1":
                find_flights(a)
            elif option == "2":
                find_flight(a)
            elif option == "3":
                reserve_flight(a)
            elif option == "4":
                cancel_flight(a)
            elif option == "5":
                add_meal(a)
            elif option == "6":
                monitor_updates(a)
            elif option == "7":
                view_reservations(a)
            elif option == "8":
                return
        except KeyboardInterrupt:
            break


if __name__ == "__main__":
    main()
