package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
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
    files, err := ioutil.ReadDir("persistence")
    if err != nil {
        return err
    }
	log.Printf("%+v\n", args.Receiver) 

    var messages []Message
    for _, file := range files {
        if filepath.Ext(file.Name()) == ".json" {
            data, err := ioutil.ReadFile(filepath.Join("persistence", file.Name()))
            if err != nil {
                return err
            }
            var msg Message
            if err := json.Unmarshal(data, &msg); err != nil {
                return err
            }
			if msg.Receiver == args.Receiver {
            	messages = append(messages, msg)
            	log.Printf("Message read: %+v\n", msg) // Print the message read from the file
			}
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