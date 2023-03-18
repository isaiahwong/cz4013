import time


class Flight:
    def __init__(
        self,
        id: int = 0,
        source: str = "",
        destination: str = "",
        airfare: float = 0.0,
        seat_availability: int = 0,
        timestamp: int = time.time() * 1000,
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
        id: int = 0,
        name: str = "",
    ):
        self.id = id
        self.name = name

    def __str__(self):
        return f"Id: {self.id}\nName: {self.name}"


class ReserveFlight:
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
