package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	gRPC "ChittyChat2.0/chat"
	"google.golang.org/grpc"
)

func main() {
	port := "localhost:5001" //connecting to same port as server
	connection, err := grpc.Dial(port, grpc.WithInsecure()) //with insecure: disables transport security
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer connection.Close() //closes connection at the end of the function

	client := gRPC.NewChatClient(connection) //creates a new client
	clientName := os.Args[1] //takes a name from the terminal

	/*Context
		- Carries deadlines, cancellation signals and other request-scoped values
		across API boundries and between processes
		- Incoming requests should create a context
		- Outgoing calls should accept a context
	*/
	cont := context.Background() 

	//Calling *THE SERVICE* from PB file, which returns a stream
	stream, err := client.Chat(cont) 
	if err != nil {
		log.Fatal(err)
	}

	//Creating a thread with an infinite loop to keep sending messages/requests
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			// Holds the string that was scanned
			text := scanner.Text()
			if len(text) != 0 {
				if len(text) > 128 {
					log.Println("Message too long")
					continue
				}
				fmt.Println(text)
			} else {
				// exit if user entered an empty string
				break
			}
			message := text
			
			/*
			- How is it calling SendMsg directly on a stream, when it is wrapped inside of a function?
			- Does &gRPC.BroadcastRequest{Message: mes} return the x.ClientStream.SendMsg(m)
			which is then passed into the SendMessage?
			*/
			//Sending message using the broadcast request "message" //technical word for message?
			if err := stream.SendMsg(&gRPC.BroadcastRequest{Name: clientName, Message: message, Time: 1}); err != nil {
				log.Fatal(err)
			}
			log.Printf("Message sent")
			//log.Printf("Sent message: %s", message)
			waitTime := rand.Intn(5)
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		// handle error
		if scanner.Err() != nil {
			log.Printf("Error: ", scanner.Err())
		}
	}()
	//Infinite loop for receiving messages
	for {
		_, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Message received by other users")
		//log.Printf("Received message from %s", response.Message)
	}

}