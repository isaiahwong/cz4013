# Running the Server
The flight data is retrieved from a `flights.csv` data file which is generated from the `scripts/seed_gen.py` file.

## Running with golang
```
$ make
# or
$ go run cmd/main.go -deadline 3 -semantic 1 -loss 0 -port 8080
```

## Running with docker
```
$ docker build  -t flight-sys .
$ docker run -p 8080:8080/udp flight-sys 
```

## Running with prebuilt binaries
```
$ ./release/flightsystem-macos -deadline 3 -semantic 1 -loss 0 -port 8080
$ ./release/flightsystem-ubuntu -deadline 3 -semantic 1 -loss 0 -port 8080
```

# Running the client and interactive mode
> Interactive mode enables configuring the server/client without flags
## Running with golang
```
# Running client
$ go run cmd/main.go -c

# Running interactive mode
$ ./release/flightsystem-macos -c
```
## Running in interactive mode
```
# Running client
$ go run cmd/main.go -i

# Running interactive mode
$ ./release/flightsystem-macos -i
```

# Flags
| Fields   | Description                                                                        |
|----------|------------------------------------------------------------------------------------|
| i        | Launch in interactive mode                                                         |
| c        |     Launch in client mode.                                                         |
| deadline | The duration where the server will wait for a request in seconds.                  |
| semantic | The type of semantic that the server will run. At-Most-Once=0, At-Least-Once=1.    |
| loss     | The server’s loss rate in percentage where it drops the packet. (For simulation).  |
| port     | The port where the server will expose its endpoint.                                |

# Directory

`cmd`: Folder for entry point for code for launching application
  1. `flight_client`:  
      Entry point for launching command line.
  
  2. `server`:  
      Entry point for launching server.
  
  3. `main.go`:    
      Entry point for launching overall flight system.

`common`: Contains utility functionality 
  1. `csv.go`  
     Csv utility for loading csv files and parsing to a particular type
  
  2. `util.go`  
     Contains various utility functions such as logger and interrupt.

`encoding`: Contains codecs for marshalling and unmarshalling
  1. `codec.go`  
      Various codecs for different data types
  
  2. `decoder.go`  
      Decoder for performing unmarshalling
  
  3. `encoder.go`  
     Encoder for performing marshalling

  4. `reader.go`  
     Buffer reader for decoder to read byte stream

`protocol`: Contains networking and server implementation for stream-oriented connection
  1. `frame.go`  
      Defines protocol frame standard format.
  
  2. `options.go`  
      Defines server options.
  
  3. `server.go`  
      Server implementation that handles overall application

  4. `session.go`  
      Session layer implementation for protocol that handles multiple stream.

  5. `stream.go`  
      stream layer implementation for protocol that handles data transfer.

`release`: Contains prebuilt binaries 

`rpc`: Contains remote prodecure call methods
  1. `errors.go`  
      Various errors
  
  2. `handlers.go`  
      RPC handlers that handle the flight application
  
  3. `message.go`  
      Contains RPC message format that is used to exchange

  4. `repo.go`  
      Contains data repo that retrieves data from mock database

  5. `router.go`  
      Routes request to different RPC handlers
  
  6. `types.go`
     Contains flight types for RPC

`flights.csv`: Flight generated data


# Building the binaries from docker
> The binaries prepared were built with docker. You may reproduce this by running the following commands
```
$ docker build -f Dockerfile.build -t flight-builder . 
$ docker run --rm -v $(pwd)/release:/release flight-builder
```

