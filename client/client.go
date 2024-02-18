package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/ygordefraga/real-time-chat/shared"
)


func main() {
	u := "ws://localhost:8080/ws"
	log.Printf("Conectando a %s...\n", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}
	defer c.Close()

	clientID := getClientIDFromInput()

	err = registerClient(c, clientID)
	if err != nil {
		log.Println("Erro ao registrar cliente:", err)
		return
	}

	go readMessages(c)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		cleanupSession(c, clientID)
		os.Exit(1)
	}()

	for {
		sendMessage(c, clientID)
	}
}

func getClientIDFromInput() string {
	fmt.Print("Digite o seu ID: ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    return scanner.Text()
}

func registerClient(c *websocket.Conn, clientID string) error {
	msg := Message{
		Text:     clientID,
		Sender:   clientID,
		Receiver: "server",
		Type:     "new_client",
		Timestamp: time.Now(),
	}

	return c.WriteJSON(msg)
}

func readMessages(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler mensagem:", err)
			return
		}

		var receivedMsg Message
		err = json.Unmarshal(message, &receivedMsg)
		if err != nil {
			log.Println("Erro ao decodificar a mensagem:", err)
			return
		}

		if receivedMsg.Type == "error" {
			log.Fatal("Erro do servidor: %s\n", receivedMsg.Text)
		} else {
			log.Printf("%s: %s\n", receivedMsg.Sender, receivedMsg.Text)
		}
	}
}

func sendMessage(c *websocket.Conn, clientID string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	message := scanner.Text()

	// Split the message text by space to extract the recipient ID
	parts := strings.SplitN(message, " ", 2)
	var recipientID string
	if len(parts) > 1 && strings.HasPrefix(parts[0], "to:") {
		recipientID = strings.TrimPrefix(parts[0], "to:")
		message = parts[1] // Remove the recipient ID prefix from the message text
	}
	
	if recipientID == "" {
		log.Println("Recipient ID not provided. Please include recipient ID in the message 'to:<id> >message>'.")
		return
	}

	msg := Message{
		Text:     message,
		Sender:   clientID,
		Receiver: recipientID,
		Type:     "chat",
		Timestamp: time.Now(),
	}

	err := c.WriteJSON(msg)
	if err != nil {
		log.Println("Erro ao enviar mensagem:", err)
		return
	}
}

func cleanupSession(c *websocket.Conn, clientID string) {
	// Send a message to the server indicating the end of the session
	msg := Message{
		Text:      "Session ended",
		Sender:    clientID,
		Receiver:  "server",
		Type:      "session_end",
		Timestamp: time.Now(),
	}
	err := c.WriteJSON(msg)
	if err != nil {
		log.Println("Erro ao enviar sinal de fim de sessão:", err)
	}
}
