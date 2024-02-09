package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Text     string `json:"text"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Type     string `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	myID      string
)

func main() {
	u := "ws://localhost:8080/ws"
	fmt.Printf("Conectando a %s...\n", u)

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

	recipientID := getRecipientIDFromInput()

	go readMessages(c)

	for {
		sendMessage(c, clientID, recipientID)
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

func getRecipientIDFromInput() string {
	fmt.Print("Digite o ID do destinatário: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
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

		fmt.Printf("%s: %s\n", receivedMsg.Sender, receivedMsg.Text)
	}
}

func sendMessage(c *websocket.Conn, clientID, recipientID string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	message := scanner.Text()

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
