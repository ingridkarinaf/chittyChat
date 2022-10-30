package main

import (
	"log"
	"net"
	"os"
	"sync"

	gRPC "ChittyChat2.0/chat"
	"google.golang.org/grpc"
)

type server struct {
	gRPC.UnimplementedChatServer
	clients map[string]gRPC.Chat_ChatServer
	mu      sync.RWMutex //what is the mutex for?
	clock   int32
}

/*
- Chat_ChatServer is a server interface for sending and receiving broadcast messages
- So here we have a server srv of type Chat_ChatServer
*/
func (s *server) addClient(clientName string, srv gRPC.Chat_ChatServer) {
	s.mu.Lock()
	defer s.mu.Unlock()         //Unlocks at the end of the function
	s.clients[clientName] = srv //Adds the server of the chat client to the list of clients
}

func (s *server) removeClient(clientName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, clientName)

	//log.Printf("%s left the chat at Lamport time %v", clientName, s.clock)
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

func (s *server) updateLamport(clientClock int32) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var maxClock int32
	if clientClock > s.clock {
		maxClock = int32(clientClock)
	} else {
		maxClock = int32(s.clock)
	}

	return maxClock + 1
}

/*
- Where was this defined? Can't find it in the proto file
- Also where is it called?
*/
func (s *server) Chat(srv gRPC.Chat_ChatServer) error {

	//We get a joining request always when a new person join the chat
	//which we send to all the chat members
	joiningRequest, err := srv.Recv()
	if err != nil {
		log.Printf("Receiving error: %v", err)
	}
	//Update time: receive join message
	s.clock = s.updateLamport(joiningRequest.Time)
	log.Println("Receive join message (server): ", s.clock)

	clientName := joiningRequest.Name
	s.addClient(clientName, srv) //Adding user to the server
	//Update time: add client
	s.clock = s.updateLamport(joiningRequest.Time)
	log.Println("Adding client at Lamport time: ", s.clock)

	defer s.removeClient(clientName)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v", err)
			os.Exit(1)
		}
	}()

	//Update time: Broadcast adding client message
	s.clock = s.updateLamport(s.clock)
	log.Println("Broadcasting adding client clock: ", s.clock)
	for _, server := range s.getClients() {
		if err := server.Send(&gRPC.BroadcastResponse{Name: joiningRequest.Name, Message: joiningRequest.Message, Time: s.clock}); err != nil {
			log.Printf("Broadcasting error: %v", err)
		}
	}

	/*
		This function is continiously checking for messages
	*/
	for {

		//Error happens here because there is nothing to receive
		response, err := srv.Recv()
		if err != nil {
			log.Printf("Receiving error: %v", err)
			break
		}

		if response.Message == "exit" {
			//Update Lamport time: receive leave message
			s.clock = s.updateLamport(response.Time)
			log.Println("Server received leaving message at Lamport time: ", s.clock)
			//Server logs client removed and updated lamport time
			s.clock = s.updateLamport(response.Time)
			log.Println("Server removes client at Lamport time: ", s.clock)

			leavingMessage := response.Name + " is leaving the chat"
			//Update Lamport time: broadcast Lamport time
			s.clock = s.updateLamport(s.clock)
			log.Println("Server broadcasted the message at Lamport time: ", s.clock)
			for _, server := range s.getClients() {
				//Using server to send broadcasting messages to all clients
				if err := server.Send(&gRPC.BroadcastResponse{Name: response.Name, Message: leavingMessage, Time: s.clock}); err != nil {
					log.Printf("Broadcasting error: %v", err)
				}

			}

			continue
		}

		s.clock = s.updateLamport(response.Time)
		log.Printf("Receive message at Lamport time %v: Message from %s : %s", s.clock, response.Name, response.Message)

		//Update time: broadcast message
		s.clock = s.updateLamport(s.clock)
		log.Println("Time when broadcasting message: ", s.clock)
		for _, server := range s.getClients() {

			//Makes sure the message is not sent to the stream where the message came from
			if server == srv {
				continue
			}
			//Using server to send broadcasting messages to all clients
			if err := server.Send(&gRPC.BroadcastResponse{Name: response.Name, Message: response.Message, Time: s.clock}); err != nil {
				log.Printf("Broadcasting error: %v", err)
			}

		}
	}
	return nil
}

func main() {
	//Establishing a connection
	port := ":9080"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()               //rename s when understanding where the &server is coming from
	gRPC.RegisterChatServer(s, &server{ //where is the "server" coming from?
		clients: make(map[string]gRPC.Chat_ChatServer), //making a map of clients
		mu:      sync.RWMutex{},
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
