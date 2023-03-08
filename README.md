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
| Flag | Length of Data | SID | RID | Data |
```
### Fields of a Frame
| Fields | Description                |
|--------|----------------------------|
| Flag   | The command flag           |
| Len    | The length of data         |
| SID    | the id of the stream       |
| RID    | the request id of the stream       |
| Data   | The data being transmitted |

### Flag
| Flag | Description                          |
|------|--------------------------------------|
| SYN  | Indicates the start of a new stream. |
| PSH  | Sends data                           |
| ACK  | Marks the end of sending a data      |
| NOP  | No operation                         |
| FIN  | Terminates the stream connection     |

## Example
This section depicts a request response between a `client` and `server`
1. `client` sends a SYNC

# CZ4013 Notes 
Server: <br>
1. The information of all flights is stored
2. Flight class: flight identifier (int), the source and destination places (variable length strings), departure time (own datastructure), airfare (float), seat availability (int)


Client: <br>
1. provides an interface for users to invoke these services. 
2. On receiving a request input from the user, the client sends the request to the server. 
3. After receiving the results from the user, the client sends the request to the client. 
4. The client then presents the results on the console to the user. 

In java, serialization is the synonym of marshalling in Java. Deserialization is the synonym of unmarshalling in Java. <br>
BUT we cannnot use any existing RMI, RPC, COBRA, Java Object serialization facilities and input/output stream in Java.


