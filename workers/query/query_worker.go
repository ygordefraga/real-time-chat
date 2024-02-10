package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
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

func (t *MessageRPCServer) ReadAllMessages(args struct{ Receiver string }, reply *[]Message) error {
    folderPath := filepath.Join("persistence", args.Receiver)
    files, err := ioutil.ReadDir(folderPath)
    if err != nil {
        // If the folder doesn't exist or cannot be read, return an empty list of messages
        if os.IsNotExist(err) {
            log.Printf("Folder does not exist: %s\n", folderPath)
            *reply = []Message{}
            return nil
        }
        // For other errors, return the error
        return err
    }
    
    var messages []Message
    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            data, err := ioutil.ReadFile(filepath.Join(folderPath, file.Name()))
            if err != nil {
                return err
            }
            var msg Message
            if err := json.Unmarshal(data, &msg); err != nil {
                return err
            }
            messages = append(messages, msg)
            log.Printf("Message read: %+v\n", msg) // Print the message read from the file
        }
    }

    *reply = messages
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