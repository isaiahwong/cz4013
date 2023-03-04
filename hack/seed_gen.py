import random
import csv
from typing import List


class Flight:
    def __init__(
        self,
        ID: int,
        source: str,
        destination: str,
        timestamp: int,
        airfare: float,
        seat_availability: int,
    ):
        self.ID = ID
        self.source = source
        self.destination = destination
        self.timestamp = timestamp
        self.airfare = airfare
        self.seat_availability = seat_availability


def generate_flight_data(num_flights: int) -> List[Flight]:
    cities = [
        "New York",
        "Los Angeles",
        "Chicago",
        "Houston",
        "Phoenix",
        "Philadelphia",
        "San Antonio",
        "San Diego",
        "Dallas",
        "San Jose",
    ]
    flights = []
    for i in range(num_flights):
        flight = Flight(
            ID=random.randint(1000, 9999),
            source=random.choice(cities),
            destination=random.choice(cities),
            timestamp=random.randint(0, 2**32 - 1),
            airfare=random.uniform(100.0, 1000.0),
            seat_availability=random.randint(0, 100),
        )
        flights.append(flight)
    return flights


def main():
    flights = generate_flight_data(50)
    with open("flights.csv", mode="w", newline="") as csv_file:
        writer = csv.writer(csv_file)
        writer.writerow(
            ["ID", "Source", "Destination", "Timestamp", "Airfare", "Seat Availability"]
        )
        for flight in flights:
            writer.writerow(
                [
                    flight.ID,
                    flight.source,
                    flight.destination,
                    flight.timestamp,
                    flight.airfare,
                    flight.seat_availability,
                ]
            )


# execute main
if __name__ == "__main__":
    main()
