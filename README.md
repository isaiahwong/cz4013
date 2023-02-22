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