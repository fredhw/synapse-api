package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/challenges-fredhw/servers/gateway/sessions"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//TODO: add a handler that upgrades clients to a WebSocket connection
//and adds that to a list of WebSockets to notify when events are
//read from the RabbitMQ server. Remember to synchronize changes
//to this list, as handlers are called concurrently from multiple
//goroutines.

//WebSocketsHandler upgrades clients to a WebSocket connection
type WebSocketsHandler struct {
	ctx      *Context
	notifier *Notifier
	upgrader *websocket.Upgrader
}

//NewWebSocketsHandler constructs a new WebSocketsHandler
func NewWebSocketsHandler(notifier *Notifier, context *Context) *WebSocketsHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &WebSocketsHandler{
		ctx:      context,
		notifier: notifier,
		upgrader: &upgrader,
	}
}

//ServeHTTP implements the http.Handler interface for the WebSocketsHandler
func (wsh *WebSocketsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("received websocket upgrade request")

	//Users must be authenticated to upgrade to a WebSocket;
	//if you get an error when retrieving the session state,
	//respond with an http.StatusUnauthorized error.
	state := &sessionState{}
	_, err := sessions.GetState(r, wsh.ctx.signingKey, wsh.ctx.sessionStore, state)
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving session state: %v", err), http.StatusUnauthorized)
		return
	}

	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go wsh.notifier.AddClient(conn)
	go wsh.notifier.start()
}

//Notifier is an object that handles WebSocket notifications
type Notifier struct {
	clients []*websocket.Conn
	eventQ  chan []byte
	mx      sync.RWMutex
}

//NewNotifier constructs a new Notifier
func NewNotifier() *Notifier {
	return &Notifier{
		clients: []*websocket.Conn{},
		eventQ:  make(chan []byte),
	}
}

//AddClient adds a new client to the Notifier
func (n *Notifier) AddClient(client *websocket.Conn) {
	log.Println("adding new WebSockets client")

	n.mx.Lock()
	n.clients = append(n.clients, client)
	n.mx.Unlock()

	tempClients := []*websocket.Conn{}

	//readPump
	for {
		if _, _, err := client.NextReader(); err != nil {
			client.Close()
			n.mx.Lock()
			for _, c := range n.clients {
				if c != client {
					tempClients = append(tempClients, c)
				}
			}
			n.clients = tempClients
			n.mx.Unlock()
			break
		}
	}
}

//Notify broadcasts the event to all WebSocket clients
func (n *Notifier) Notify(event []byte) {
	log.Printf("adding event to the queue")
	n.eventQ <- event
}

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections. If you get an error
//writing to the WebSocket, just close it and remove it
//from the list (client went away without closing from
//their end).
//Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket

//start starts the notification loop
func (n *Notifier) start() {
	log.Println("starting notifier loop")

	//rabbitmq
	mqAddr := os.Getenv("MQADDR")
	if len(mqAddr) == 0 {
		mqAddr = "localhost:5672"
	}
	mqURL := fmt.Sprintf("amqp://%s", mqAddr)
	conn, err := amqp.Dial(mqURL)
	if err != nil {
		log.Fatalf("error connecting to RabbitMQ: %v", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error creating channel: %v", err)
	}
	q, err := channel.QueueDeclare("testQ", false, false, false, false, nil)

	msgs, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("error consuming messages: %v", err)
	}

	go func() {
		for msg := range msgs {
			log.Println(string(msg.Body))
			n.Notify([]byte(msg.Body))
		}
	}()

	for {
		event := <-n.eventQ
		log.Printf("event: %v", event)
		n.mx.RLock()
		for _, conn := range n.clients {
			if err := conn.WriteMessage(websocket.TextMessage, event); err != nil {
				log.Println(err)
				return
			}
		}
		n.mx.RUnlock()
	}
}
