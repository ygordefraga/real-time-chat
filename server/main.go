package main

import (
	"log"
	"net/http"
	"net/rpc"
	"time"

	"github.com/gorilla/websocket"
)

// Estrutura para manter informações da mensagem
type Message struct {
	Text     string `json:"text"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Type 	 string `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	// O Upgrader é usado para atualizar uma conexão HTTP regular para uma conexão WebSocket.
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true }, // Uma função que determina se o cliente está autorizado a fazer upgrade da conexão HTTP para WebSocket
	}
	chat_broadcast = make(chan Message) // Este é um canal que será usado para transmitir mensagens para todos os clientes conectados ao servidor WebSocket.
	persist_broadcast = make(chan Message) // Este é um canal que será usado para transmitir mensagens para todos os clientes conectados ao servidor WebSocket.
	historical_broadcast = make(chan Message)
	clients   = make(map[string]*websocket.Conn) // Este é um mapa que associa os IDs dos clientes aos objetos websocket.Conn (ponteiro). Ele é usado para rastrear todas as conexões WebSocket ativas no servidor.)
)


func handleConnections(w http.ResponseWriter, r *http.Request) {
	/*
		A função Upgrade do upgrader faz exatamente isso. Ela atualiza uma conexão HTTP normal para uma conexão WebSocket. Essa função aceita três argumentos:

		w: O objeto http.ResponseWriter usado para escrever a resposta HTTP.
		r: O objeto http.Request que representa a requisição HTTP recebida.
		nil: Opcionalmente, um objeto http.Header que pode conter cabeçalhos adicionais para a atualização.
	*/
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Essa é a parte do código responsável por ler mensagens do cliente
	// recém-conectado (conn.ReadJSON(&msg)) e enviá-las para o canal
	// chat_broadcast para que possam ser distribuídas a outros clientes conectados.
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}

		if msg.Type == "new_client" {
			if addNewUser(msg, conn) {
				go readAllMessagesFromRPC(msg.Sender)
			}
		} else if msg.Type == "chat" {
			chat_broadcast <- msg
			persist_broadcast <- msg
		} else {
			log.Println("Unknown message:", msg.Type)
		}
	}
}

func addNewUser(msg Message, conn *websocket.Conn) bool {
	// Adicione o cliente ao mapa de clientes
	clientID := msg.Text

	// Check if the client ID is already in use
	if _, exists := clients[clientID]; exists {
		// Client ID already in use, return an error
		log.Printf("User already exists: %s\n", clientID)
		err := conn.WriteJSON(Message{
			Text:     "User already exists",
			Sender:   "server",
			Receiver: clientID,
			Type:     "error",
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Println("Erro ao enviar mensagem de erro:", err)
		}
		return false
	} else {
		clients[clientID] = conn
		log.Printf("New user: %s\n", clientID)
		return true
	}

}

func handleChatMessages() {
    for {
        msg := <- chat_broadcast

		log.Printf("Enviando mensagem do cliente %s para o cliente %s: %s\n", msg.Sender, msg.Receiver, msg.Text)
		// Verifique se o destinatário está online e envie a mensagem apenas para ele
		if conn, ok := clients[msg.Receiver]; ok {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				conn.Close()
				delete(clients, msg.Receiver)
			}
		}
    }
}

func handleHistoricalMessages() {
    for {
        msg := <- historical_broadcast

		log.Printf("Enviando mensagem do cliente %s para o cliente %s: %s\n", msg.Sender, msg.Receiver, msg.Text)
		// Verifique se o destinatário está online e envie a mensagem apenas para ele
		if conn, ok := clients[msg.Receiver]; ok {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				conn.Close()
				delete(clients, msg.Receiver)
			}
		}
    }
}

func forwardMessagesToRPC() {
	client, err := rpc.DialHTTP("tcp", "localhost:1123") // Assuming RPC server is running on localhost:1122
    if err != nil {
        log.Fatal("error connecting to RPC server:", err)
    }
    defer client.Close()

    for {
        msg := <- persist_broadcast
        var reply string
        err := client.Call("MessageRPCServer.PersistMessage", msg, &reply)
        if err != nil {
            log.Fatal("error calling RPC service:", err)
        }

        log.Println("RPC service response:", reply)
    }
}

func readAllMessagesFromRPC(receiver string) {
	client, err := rpc.DialHTTP("tcp", "localhost:1122") // Assuming RPC server is running on localhost:1122
    if err != nil {
        log.Fatal("error connecting to RPC server:", err)
    }
    defer client.Close()

    var reply []Message
	args := struct{ Receiver string }{Receiver: receiver} // Create and initialize the struct
    err = client.Call("MessageRPCServer.ReadAllMessages", args, &reply)
    if err != nil {
        log.Fatal("error calling RPC service:", err)
    }

    // Process the reply, which contains all messages
    for _, msg := range reply {
		historical_broadcast <- msg
    }
}


func main() {
	// Configuração de rotas
	http.HandleFunc("/ws", handleConnections)

	// Inicia o servidor
	log.Println("Servidor iniciado na porta 8080")

	go handleHistoricalMessages()
	go handleChatMessages()
	go forwardMessagesToRPC()
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor: ", err)
	}
}
