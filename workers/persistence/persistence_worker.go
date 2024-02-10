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

    // Create a folder for the receiver if it doesn't exist
    folderName := "persistence/" + msg.Receiver
    if _, err := os.Stat(folderName); os.IsNotExist(err) {
        if err := os.MkdirAll(folderName, 0755); err != nil {
            return err
        }
    }

    // Create a file with a unique name based on timestamp inside the receiver's folder
    filename := folderName + "/message_" + msg.Sender + "_" + msg.Timestamp.Format("2006_01_02_15_04_05") + ".json"
    file, err := os.Create(filename)
    if err != nil {
        log.Printf("Error creating file: %v", err)
        return err
    }
    defer file.Close()

    // Encode message to JSON and write to file
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(msg); err != nil {
        return err
    }

    *reply = "Message Persisted"
    return nil
}

func main() {
    // create and register the rpc
    messageRPC := new(MessageRPCServer)
    rpc.Register(messageRPC)
    rpc.HandleHTTP()

    // set a port for the server
    port := ":1123"

    // listen for requests on 1122
    listener, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal("listen error: ", err)
    }

    http.Serve(listener, nil)
}