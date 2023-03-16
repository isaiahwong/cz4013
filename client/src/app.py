import asyncio
from protocol import Client, Stream
from queue import Queue, Empty
from misc import futuretime
from threading import Thread
import time
from flight import Flight, ReserveFlight, Food
from message import message, ErrorMsg, Message, Error
from frame import EOF
import codec
import datetime


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

    def on_enter_quit(self):
        input("\nPress enter to quit..\n")
        return EOF("")

    def _sendOnly(
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
            stream = self._sendOnly(existing, method, query, deadline)
            try:
                b = stream.read()
                msg: Message = codec.unmarshal(b, Message())
            except EOF as e:
                pass
            if msg.error:
                raise Exception(msg.error)
            return [stream, msg]

        tries = 0
        while tries < self.retries:
            try:
                stream, msg = retrySend(stream)
                return [stream, msg]
            except Exception as e:
                print(e)
                print(f"Retrying: {tries}\n")
                tries += 1

        raise Exception(f"Failed to send {method} after {self.retries} tries")

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

    def get_meals(self):
        method = "GetMeals"

        req = Message(
            rpc=method,
        )

        stream, msg = self._send(method, req.query, self.deadline)
        f = codec.unmarshal(msg.body, [Food()])
        stream.close()
        return f

    def add_meals(self, id: str, meal_id: str):
        method = "AddMeals"
        req = Message(
            rpc=method,
            query={
                "id": id,
                "meal_id": meal_id,
            },
        )

        stream, msg = self._send(method, req.query, self.deadline)
        rf = codec.unmarshal(msg.body, ReserveFlight())
        stream.close()
        return rf

    def monitor_updates(self, duration: int, blocking=True):
        method = "MonitorUpdates"
        query = {"timestamp": str(int(futuretime(duration * 60 * 60) * 1000))}
        stream = self._sendOnly(None, method, query, None)

        def read():
            while not stream.closed:
                b = None
                try:
                    b = stream.read()
                except EOF as e:
                    return
                res: Message = codec.unmarshal(b, Message())
                if res.error:
                    raise Exception(res.error)

                flight: Flight = codec.unmarshal(res.body, Flight())
                print("New Flight")
                print("Flight Id: ", flight.id)
                print("Source: ", flight.source)
                print("Destination: ", flight.destination)
                print("Airfare: ", flight.airfare)
                print("Seats Reserved: ", flight.seat_availability)

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
                    async_wrap(self.on_enter_quit),
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
