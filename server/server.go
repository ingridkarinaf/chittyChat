package main

import (
	"log"
	"net"
	"os"
	"sync"

	gRPC "ChittyChat2.0/chat"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type server struct {
	gRPC.UnimplementedChatServer
	clients map[string]gRPC.Chat_ChatServer
	mu      sync.RWMutex //what is the mutex for?
}

/*
- Chat_ChatServer is a server interface for sending and receiving broadcast messages
- So here we have a server srv of type Chat_ChatServer
*/
func (s *server) addClient(userId string, srv gRPC.Chat_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()     //Unlocks at the end of the function
	s.clients[userId] = srv //Adds the server of the chat client to the list of clients
}

func (s *server) removeClient(userId string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, userId)
	log.Printf("Client left the chat")
}

func (s *server) getClients() []gRPC.Chat_ChatServer {
	var clientServers []gRPC.Chat_ChatServer //creating a list to contain client servers

	//RLock() multiple go routines can read from the lock but not write
	s.mu.RLock()
	defer s.mu.RUnlock()

	//Goes through clients in server and adds them to a new list
	//Why not just return s.clients?
	for _, c := range s.clients {
		clientServers = append(clientServers, c)
	}
	return clientServers
}

/*
- Where was this defined? Can't find it in the proto file
- Also where is it called?
*/
func (s *server) Chat(srv gRPC.Chat_ChatServer) error {
	userId := uuid.Must(uuid.NewRandom()).String() //Generating a random ID for a user

	log.Printf("New User: %s", userId)

	s.addClient(userId, srv)     //Adding user to the server
	defer s.removeClient(userId) //removing client at the end of the function

	/*
		- Executing at the end of the function
		- ?? Does it exit if there's an error, or what is this for?
	*/
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v", err)
			os.Exit(1)
		}
	}() //"defer function must be in a function call - what is this syntax?

	//We get a joining request always when a new person join the chat
	//which we send to all the chat members
	joiningRequest, err := srv.Recv()
	if err != nil {
		log.Printf("Receiving error: %v", err)
	}
	for _, server := range s.getClients() {
		if err := server.Send(&gRPC.BroadcastResponse{Name: joiningRequest.Name, Message: joiningRequest.Message}); err != nil {
			log.Printf("Broadcasting error: %v", err)
		}
	}

	/*
		This function is continiously checking for messages
	*/
	for {
		response, err := srv.Recv()

		if err != nil {
			log.Printf("Receiving error: %v", err)
			break
		}

		log.Printf("broadcast: Message from %s : %s", response.Name, response.Message)
		for _, server := range s.getClients() {

			//Makes sure the message is not sent to the stream where the message came from
			if server == srv {
				continue
			}
			//Using server to send broadcasting messages to all clients
			if err := server.Send(&gRPC.BroadcastResponse{Name: response.Name, Message: response.Message}); err != nil {
				log.Printf("Broadcasting error: %v", err)
			}

		}
	}
	return nil
}

func main() {
	//Establishing a connection
	/*
		- How come it is listening to the port before creating a server?
		- Isn't it the server that is listening to the port?
		- Although in the client, the main function established the connection,
		and then creates multiple clients
	*/
	port := ":5001"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	/*
		- Here they register with the clients inside, whereas in another example (suvi's) they don't.
		-
	*/
	s := grpc.NewServer()               //rename s when understanding where the &server is coming from
	gRPC.RegisterChatServer(s, &server{ //where is the "server" coming from?
		clients: make(map[string]gRPC.Chat_ChatServer), //making a map of clients
		mu:      sync.RWMutex{},
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
