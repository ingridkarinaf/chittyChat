<!-- # The largest heading
## The second largest heading
###### The smallest heading -->

# Distributed Systems
## Mandatory Assignment - ChittyChat

### Set-up 
1. Build a docker image: 
`docker build -t test --no-cache .`
2. Run the server in a docker container: 
`docker run -p 9080:9080 -tid test`
3. Run clients in terminal: 
`go run client/client.go nameOfChatParticipant`

### System Architecture
We're using a server-client architecture, along with bidirectional streaming, in which both the server and the client is sending and receiving data.

### RPC Methods
<!-- Describe what RPC methods are implemented, of what type, and what messages types are used for communication -->
We have a `Chat` rpc method of type **stream** which has messages `broadcastRequest` and `broadcastResponse`. The client is able to sent a join request, send messages for other members and leave the chat by typing "exit" within the Chat method. 

### Lamport Timestamp Implementation
<!-- - Describe how you have implemented the calculation of the Lamport timestamps -->
The following are considered events that increments the Lamport timestamp (each increments by 1): 
- Join chat (client)
- Adding a client (server)
- Removing a client (server)
- Sending a message (client and server)
- Receiving a message (client and server)
- Broadcasting, i.e. sending messages to all clients (server)
- Leave chat (client)

As each of these events are happening, the relevant node will compare their own clock with the node it communicates with, take the max, increment the max value by 1, and assign the value to its own clock. 

<!-- - Provide a diagram, that traces a sequence of RPC calls together with the Lamport
timestamps, that corresponds to a chosen sequence of interactions: Client X joins, Client X Publishes, ..., Client X leaves. Include documentation (system logs) in your appendix. -->
Go to link below for lamport time demo (diagram and terminal outputs):
https://github.com/ingridkarinaf/chittyChat/tree/main/Demonstrating_lamport_time

<!-- - Provide a link to a Git repo with your source code in the report -->
<!-- - Include system logs, that document the requirements are met, in the appendix of
your report -->

