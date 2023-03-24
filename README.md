# Quick start
## Running Golang's server/client with prebuilt binaries
Prebuilt binaries have been created in the `server/release folder`. 
Run the appropriate binary for the respective OS

## Running the server
```
# macOS
$ chmod +x ./server/release/flightsystem-macos 
$ ./server/release/flightsystem-macos

# linux
$ chmod +x ./server/release/flightsystem-ubuntu
$ ./server/release/flightsystem-ubuntu

# windows
$ ./server/release/flightsystem-windows.exe
```

## Running the clent
Open another terminal and run the client
```
# macOS
$ ./server/release/flightsystem-macos -c

# linux
$ ./server/release/flightsystem-ubuntu -c

# windows
$ ./server/release/flightsystem-windows.exe -c
```

## Running with docker
> Note the docker image runs the server in it's default config
> Running the client on docker is not recommended as you have to somehow pass in stdin
```
$ docker build  -t flight-sys .
$ docker run -p 8080:8080/udp flight-sys
```

## Running in interactive mode
> Interactive mode provides user friendly prompts to run the server
```
# macOS
$ ./server/release/flightsystem-macos -i

# linux
$ ./server/release/flightsystem-ubuntu -i

# windows
$ ./server/release/flightsystem-windows.exe -i
```

## Running Golang's server/client from src
[Installation of golang](https://go.dev/doc/install)
### Download dependencies 
```
# Download deps
$ cd server
$ go mod download

$ make
# or
$ go run cmd/main.go -i
```

## Running Python's client from src
```
$ cd client

$ make
# or
$ python3 src/main.py

```

# Folder directory
`client`: Python client implementation of flight system
`scripts`: Contains scripts that help generate dummy flight data
  1. `flights.csv` - Generated flight.csv
  2. `seed_gen.py` - Python script to generate flight data
   
`server`: Golang server and client implementation of flight system
  - Refer to the server's readme for more details

# Generate flight data
> Note the make command generates the `flight.csv` file and outputs to the server
```
$ make 
# or
$ python3 srcipts/seed_gen.py && mv flights.csv ./server
```



# Protocol
The project utilises UDP as its transport layer where the communication is further decomposed into multiple layers. The byte ordering used is `Little Endian`.

## Session
A single instance of server running will maintain one session in a single UDP port. A session can have multiple `streams` where it will multiplex requests. 

## Stream
A stream is typically opened when a client makes a request to the server. A `Frame` is the data standard when communicating in a stream. A `Frame` consists of a `flag` that decicates how the stream should handle the data transimission. 

## Frame
Frame is the mode of data encapsulation. A frame in transmission will have a `header` and `body`

<br/>
A frame in transmission would have the byte arrangement as such:

```
| Flag | Length of Data | SID | RID | SeqID | Data |
```
### Fields of a Frame
| Fields | Description                |
|--------|----------------------------|
| Flag   | The command flag           |
| Len    | The length of data         |
| SID    | The id of the stream       |
| RID    | The request id of the stream       |
| SeqID  | The sequence of a frame that may have been broken down |
| Data   | The data being transmitted |

### Flag
| Flag | Description                          |
|------|--------------------------------------|
| SYN  | Indicates the start of a new stream. |
| PSH  | Sends data                           |
| DNE  | Marks the end of sending a data      |
| NOP  | No operation                         |
| FIN  | Terminates the stream connection     |

