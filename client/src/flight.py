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


class Food:
    def __init__(
        self,
        id: int = 0,
        name: str = "",
    ):
        self.id = id
        self.name = name


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
        return f"ReserveFlight(id={self.id}, flight={self.flight}, seats_reserved={self.seats_reserved}, check_in={self.check_in}, cancelled={self.cancelled}, meals={self.meals})"
