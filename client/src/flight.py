import time


class Flight:
    """
    In this project, there are three main objects to consider:
    1. Flight
    2. Food
    3. Reserve Flight
    This python3 file declares these three as classes and with their respective attributes
    When the RPC is called, it can return any of these three object types
    """

    def __init__(
        self,
        id: int = 0,  # The Unique ID of the flight
        source: str = "",  # The origin of the flight
        destination: str = "",  # The destination of the flight
        airfare: float = 0.0,  # The cost of the flight
        seat_availability: int = 0,  # The number of seats available
        timestamp: int = time.time() * 1000,  # The time of departure
    ):
        self.id = id
        self.source = source
        self.destination = destination
        self.airfare = airfare
        self.seat_availability = seat_availability
        self.timestamp = timestamp

    def __str__(self):
        return f"Flight Id: {self.id}\nSource: {self.source}\nDestination: {self.destination}\nAirfare: {self.airfare}\nSeat Available: {self.seat_availability}\n"


class Food:
    def __init__(
        self,
        id: int = 0,  # The unique ID of the food
        name: str = "",  # The name of the food
    ):
        self.id = id
        self.name = name

    def __str__(self):
        return f"Id: {self.id}\nName: {self.name}"


class ReserveFlight:
    """
    Representation of a reservation gof a flight
    flight - The number of seats reserved by the user using the client
    seats_reserved - Whether the user using the client has checked in
    check_in - Whether the reserved seat has been cancelled or not
    cancelled - If a reservation is cancelled
    meals - The list of foods the user using the client has pre orders for the flight
    """

    def __init__(
        self,
        id: str = "",
        flight: Flight = Flight(),
        seats_reserved: int = 0,
        check_in: bool = False,
        cancelled: bool = False,
        meals: list = [Food()],
    ):
        self.id = id
        self.flight = flight
        self.seats_reserved = seats_reserved
        self.check_in = check_in
        self.cancelled = cancelled
        self.meals = meals

    def __str__(self):
        s = f"Flight Id: {self.flight.id}\nSource: {self.flight.source}\nDestination: {self.flight.destination}\nAirfare: {self.flight.airfare}\nSeats Reserved: {self.seats_reserved}\nCancelled: {self.cancelled}"
        m = [meal.name for meal in self.meals]
        return s + f"\nMeals: {m}\n"
