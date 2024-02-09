package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

// an RPC server in Go

type Message struct {
    Text     string `json:"text"`
    Sender   string `json:"sender"`
    Receiver string `json:"receiver"`
    Type     string `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

type MessageRPCServer string

func (t *MessageRPCServer) PersistMessage(msg Message, reply *string) error {
    log.Printf("Received message: %s from sender: %s, receiver: %s\n", msg.Text, msg.Sender, msg.Receiver)

	filename := "persistence/message_" + msg.Sender + "_" + msg.Receiver + "_" + msg.Timestamp.Format("2006_01_02_15_04_05") + ".json"
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
	}
	defer file.Close()
	
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(msg); err != nil {
        return err
    }

	*reply = "Message Persisted"

    return nil
}

func saveMessageToFile(msg Message) error {
    // Create a file with a unique name based on timestamp
    filename := "message_" + msg.Sender + "_" + msg.Receiver + ".json"
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // Encode message to JSON and write to file
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(msg); err != nil {
        return err
    }

    log.Printf("Message saved to file: %s\n", filename)
    return nil
}

func main() {
    // create and register the rpc
    messageRPC := new(MessageRPCServer)
    rpc.Register(messageRPC)
    rpc.HandleHTTP()

    // set a port for the server
    port := ":1122"

    // listen for requests on 1122
    listener, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal("listen error: ", err)
    }

    http.Serve(listener, nil)
}