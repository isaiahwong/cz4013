import asyncio
from protocol import Client, Stream
from queue import Queue, Empty
from misc import futuretime
from threading import Thread, Event
import time
from flight import Flight, ReserveFlight, Food
from message import Message, Error
from frame import EOF
import codec
import datetime
import select
import sys


#Checks if the platform is a windows platform 
if sys.platform.startswith("win"):
    import msvcrt

'''
The App Class:
This class allows the client to invoke the RPCs provided by the server:
1. FindFlights: Returns available flights when provided with a origin and destination of flight
2. FindFlight: Returns the available flight when provided with the flight id
3. ReserveFlight: Reserves a desired number of seats on a flight if available. Returns an error if seats are not available 
4. MonitorUpdates: Monitors flight reservations and cancellations
5. CancelFlights: Cancels a flight reservations
6. AddMeals: Add meals for a reserved flight 
'''


class App:
    def __init__(self, remote="127.0.0.1", port=8080, retries=2, deadline=1, mtu=1500):
        self.remote = remote
        self.port = port
        self.retries = retries
        self.deadline = deadline
        self.mtu = mtu
        self.reservations = {}
        self.client = Client(remote, port)
        self.keyEnter = Queue()
        self.stop_event = Event()

    """ Helper function to return all the reservation ids  """
    def reservations_idx(self):
        return [v for k, v in self.reservations.items()]

    """ Helper function to print all the reservation ids """
    def print_reservations(self):
        for i, v in enumerate(self.reservations_idx()):
            print(f"Reservation[{i}]\n{v}\n")

    """ A function to call the FindFlights RPC and simply return the flights """
    def find_flights(self, source, destination) -> list:
        method = "FindFlights"
        stream = None
        if not source:
            raise Exception(f"{method}: Invalid source")
        if not destination:
            raise Exception(f"{method}: Invalid destination")

        req = Message(
            rpc=method,
            query={"source": source, "destination": destination},
        )

        stream, msg = self._send(method, req.query, self.deadline)
        flights = codec.unmarshal(msg.body, [Flight()])
        stream.close()
        return flights


    """ A function to call the FindFlight RPC and simple return the flight (unique) """
    def find_flight(self, id: str) -> Flight:
        method = "FindFlight"
        stream = None
        if not id:
            raise Exception(f"{method}: Invalid id")

        req = Message(
            rpc=method,
            query={"id": id},
        )

        stream, msg = self._send(method, req.query, self.deadline)

        flight = codec.unmarshal(msg.body, Flight())
        stream.close()
        return flight

    """ A function to call the ReserveFlight RPC and returns the reserved Flight and ID """
    def reserve_flight(self, id: str, seats: str) -> ReserveFlight:
        method = "ReserveFlight"
        stream = None
        if not id:
            raise Exception(f"{method}: Invalid id")
        if not seats:
            raise Exception(f"{method}: Invalid seats")

        req = Message(
            rpc=method,
            query={"id": id, "seats": seats},
        )

        stream, msg = self._send(method, req.query, self.deadline)
        r = codec.unmarshal(msg.body, ReserveFlight())
        self.reservations[r.id] = r
        stream.close()
        return r

    """ A function to call the CancelFlight RPC and returns the reserved Flight """
    def cancel_flight(self, id: str) -> ReserveFlight:
        if len(self.reservations) == 0:
            return None

        method = "CancelFlight"
        stream = None
        if not id:
            raise Exception(f"{method}: Invalid id")

        if not id in self.reservations:
            raise Exception(f"{method}: Reservation not found")

        req = Message(
            rpc=method,
            query={"id": id},
        )

        stream, msg = self._send(method, req.query, self.deadline)

        r = codec.unmarshal(msg.body, ReserveFlight())
        del self.reservations[r.id]
        stream.close()
        return r

    """ A function that calls the GetMeals RPC and returns the Food available on that flight """
    def get_meals(self):
        method = "GetMeals"

        req = Message(
            rpc=method,
        )

        stream, msg = self._send(method, req.query, self.deadline)
        f = codec.unmarshal(msg.body, [Food()])
        stream.close()
        return f

    """ A function that calls the AddMeals RPC and returns the reserved Flight in which the meal is booked """
    def add_meals(self, id: str, meal_id: str):
        method = "AddMeals"
        req = Message(
            rpc=method,
            query={
                "id": str(id),
                "meal_id": str(meal_id),
            },
        )

        stream, msg = self._send(method, req.query, self.deadline)
        rf = codec.unmarshal(msg.body, ReserveFlight())
        stream.close()
        return rf


    """ A function that calls the MonitorUpdates RPC and shows any changes to reserved flights """
    def monitor_updates(self, duration: int, blocking=True):
        method = "MonitorUpdates"
        query = {"timestamp": str(int(futuretime(duration * 60) * 1000))}
        stream = self._send_only(None, method, query, None)

        def read():
            while not stream.closed:
                b = None
                try:
                    b = stream.read()
                    res: Message = codec.unmarshal(b, Message())
                    if res.error:
                        raise Exception(res.error)

                    flight: Flight = codec.unmarshal(res.body, Flight())
                    print(f"{flight}\n")
                except EOF as e:
                    if self.stop_event:
                        self.stop_event.set()
                    return

        async def async_wrap(fn):
            """Wraps a function in to an async function"""
            loop = asyncio.get_running_loop()
            future = loop.run_in_executor(None, fn)
            result = await future
            return result

        async def concurrent():
            """Executes async functions concurrently"""
            done, pending = await asyncio.wait(
                [
                    async_wrap(read),
                    async_wrap(self._on_enter_quit),
                ],
                return_when=asyncio.FIRST_COMPLETED,
            )
            stream.close()

        if blocking:
            asyncio.run(concurrent())
            return

        t = Thread(target=read)
        t.daemon = True
        t.start()

    def _on_enter_quit(self):
        self.stop_event = Event()
        print("\nPress enter to quit..\n")
        enter_received = False
        while not self.stop_event.is_set():
            if sys.platform.startswith("win"):
                enter_received = True if msvcrt.kbhit() else False
            else:
                enter_received, _, _ = select.select([sys.stdin], [], [], 1)

            if enter_received:
                return EOF("")

        return EOF("")

    def _send_only(
        self, existing: Stream, method: str, query: dict, deadline: int = None
    ) -> Stream:
        stream = (
            self.client.open(deadline)
            if not existing
            else self.client.openWithExisting(stream, deadline)
        )
        req = Message(rpc=method, query=query)
        stream.write(codec.marshal(req))
        return stream

    def _send(self, method: str, query: dict, deadline: int):
        stream = None
        msg = None

        def retrySend(existing: Stream):
            msg = None
            stream = self._send_only(existing, method, query, deadline)
            try:
                b = stream.read()
                msg: Message = codec.unmarshal(b, Message())
            except EOF as e:
                pass
            if msg.error:
                raise Exception(f"{msg.error.error}: {msg.error.body}")
            return [stream, msg]

        tries = 0
        while tries < self.retries:
            try:
                stream, msg = retrySend(stream)
                return [stream, msg]
            except TimeoutError as e:
                print(f"Retrying: {tries}\n")
                tries += 1

        raise Exception(f"Failed to send {method} after {self.retries} tries")
