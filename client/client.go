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

var clock int32

func main() {

	clock = 0
	port := "localhost:9080"                                //connecting to same port as server
	connection, err := grpc.Dial(port, grpc.WithInsecure()) //with insecure: disables transport security
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}

	client := gRPC.NewChatClient(connection) //creates a new client
	clientName := os.Args[1]                 //takes a name from the terminal

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

	//Happens when the client exits the chat
	// defer func() {
	// 	fmt.Println("firing defer function")
	// 	clock = updateLamport(clock)
	// 	stream.SendMsg(&gRPC.BroadcastRequest{Time: int32(clock)})
	// 	if err := recover(); err != nil {
	// 		log.Printf("panic: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	fmt.Println("end of function")
	// }()

	defer connection.Close() //closes connection at the end of the function
	joiningMessage := "joined ChittyChat"
	clock = updateLamport(clock)
	log.Println("joinChat clock: ", clock)
	stream.SendMsg(&gRPC.BroadcastRequest{Name: clientName, Message: joiningMessage, Time: int32(clock)})

	//Creating a thread with an infinite loop to keep sending messages/requests
	exit := false
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for {
			scanner.Scan()
			// Holds the string that was scanned
			text := scanner.Text()
			if text == "exit" {
				exit = true
				clock = updateLamport(clock)
				stream.SendMsg(&gRPC.BroadcastRequest{Name: clientName, Message: text, Time: int32(clock)})
				break
			} else if len(text) != 0 && len(text) < 128 {
				log.Println("if statement true")
				fmt.Println(text)
			} else {
				// exit if user entered an empty string
				log.Println("Message empty or too long")
				continue
			}
			message := text

			/*
				- How is it calling SendMsg directly on a stream, when it is wrapped inside of a function?
				- Does &gRPC.BroadcastRequest{Message: mes} return the x.ClientStream.SendMsg(m)
				which is then passed into the SendMessage?
			*/
			//Update time: send message
			clock = updateLamport(clock)
			log.Println("Message sent at Lamport time: ", clock)
			//Sending message using the broadcast request "message" //technical word for message?
			if err := stream.SendMsg(&gRPC.BroadcastRequest{Name: clientName, Message: message, Time: int32(clock)}); err != nil {
				log.Fatal(err)
			}

			//log.Printf("Sent message: %s", message)
			waitTime := rand.Intn(5)
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		// handle error
		// if scanner.Err() != nil {
		// 	log.Printf("Error: %s", scanner.Err())
		// }

	}()
	//Infinite loop for receiving messages
	for !exit {
		response, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		//Update time: receive message
		clock = updateLamport(response.Time)

		if response.Message == joiningMessage {
			log.Printf("%s %s at Lamport time %v", response.Name, response.Message, clock)
		} else {
			log.Printf("Message from %s at Lamport time %v: %s", response.Name, clock, response.Message)

		}

	}

}
func updateLamport(responseClock int32) int32 {

	maxClock := int32(0)
	if responseClock > clock {
		maxClock = int32(responseClock)
	} else {
		maxClock = int32(clock)
	}

	return maxClock + 1
}
