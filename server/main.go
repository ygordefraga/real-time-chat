package main

import (
	"fmt"
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
		fmt.Println(err)
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
			fmt.Println(err)
			return
		}

		if msg.Type == "new_client" {
			handleNewUserMessages(msg, conn)
		} else if msg.Type == "chat" {
			chat_broadcast <- msg
			persist_broadcast <- msg
		} else {
			fmt.Println("Tipo de mensagem desconhecido:", msg.Type)
		}
	}
}

func handleNewUserMessages(msg Message, conn *websocket.Conn) string {
	// Adicione o cliente ao mapa de clientes
	clientID := msg.Text
	clients[clientID] = conn
	log.Printf("Novo usuário adicionado: %s\n", clientID)
	return clientID
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

func forwardMessagesToRPC() {
	client, err := rpc.DialHTTP("tcp", "localhost:1122") // Assuming RPC server is running on localhost:1122
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


func main() {
	// Configuração de rotas
	http.HandleFunc("/ws", handleConnections)

	// Inicia o servidor
	log.Println("Servidor iniciado na porta 8080")
	go handleChatMessages()
	go forwardMessagesToRPC()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor: ", err)
	}
}
