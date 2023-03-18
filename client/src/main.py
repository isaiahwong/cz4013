from protocol import Client, Stream
import time
from flight import Flight, ReserveFlight
from message import Message, Error
import codec
import datetime
from misc import futuretime
from app import App

# from threading import Thread, Event
from multiprocessing import Process, Event

# Done with addition of error checking
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


def find_flight(app: App):
    id = str(input("Enter Flight id: "))

    try:
        f = app.find_flight(id)
        print(f)
    except Exception as e:
        print(f"Error: {e}\n")


def reserve_flight(app: App):
    flightid = input("\nEnter Flight id: ")
    seats = input("\nEnter number of seats: ")
    try:
        r: ReserveFlight = app.reserve_flight(flightid, seats)
        print(f"{r}\n")
    except Exception as e:
        print(f"Error: {e}\n")


def cancel_flight(app: App):
    if len(app.reservations) == 0:
        print("You have no reservations")
        return

    print("\nReservations:\n")
    app.print_reservations()
    idxToReservations = app.reservations_idx()
    idx = -1
    while idx < 0 or idx >= len(idxToReservations):
        idx = int(input("Select reservation: "))

    try:
        rf = app.cancel_flight(idxToReservations[idx].id)
        if not rf:
            return
        f = rf.flight
        print("Reservation cancelled")
    except Exception as e:
        print(f"Error: {e}\n")


def view_reservations(app: App):
    if len(app.reservations) == 0:
        print("\nYou have no reservations\n")
        return
    print()
    app.print_reservations()


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
            meal_idx = int(input("Select meal: "))

        # Submit RPC request
        rf = app.add_meals(reservations[r_idx].id, meals[meal_idx].id)
        if not rf:
            return
        print(rf)
        print("\nMeal Added\n")
    except Exception as e:
        print(f"Error: {e}\n")


def monitor_updates(app: App):
    try:
        duration = int(input("Monitor updates Duration (minutes): "))
        app.monitor_updates(duration, blocking=True)
    except Exception as e:
        print(f"Error: {e}\n")


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


def clientFactory():
    """Client factory"""
    d = input("Load default Y/N: ").upper()
    if d != "N":
        return App()

    addr = input("Enter remote address: ").split(":")
    remote = addr[0]
    port = int(addr[1])
    return App(remote=remote, port=port)


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
