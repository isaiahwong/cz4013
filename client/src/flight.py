import time


class Flight:
    def __init__(
        self,
        id: int = 0,
        source: str = " ",
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


class ReserveFlight:
    def __init__(
        self, id: str = "", flight: Flight = Flight(), seats_reserved: int = 0
    ):
        self.id = id
        self.flight = flight
        self.seats_reserved = seats_reserved

        
