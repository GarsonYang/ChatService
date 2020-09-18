package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-GarsonYang/servers/gateway/sessions"
)

type Notifier struct {
	ConnectionMap map[int64][]*websocket.Conn
	mx            sync.RWMutex
}

func (n *Notifier) InsertConnection(userID int64, conn *websocket.Conn) int {
	n.mx.Lock()
	connID := len(n.ConnectionMap[userID])
	n.ConnectionMap[userID] = append(n.ConnectionMap[userID], conn)
	n.mx.Unlock()
	return connID
}

func (n *Notifier) RemoveConnection(userID int64, connID int) {
	n.mx.Lock()
	n.ConnectionMap[userID] = append(n.ConnectionMap[userID][:connID], n.ConnectionMap[userID][connID+1:]...)
	n.mx.Unlock()
}

func (n *Notifier) WriteToConnections(data *event, userIDs []int64) error {
	if len(userIDs) == 0 {
		for _, conns := range n.ConnectionMap {
			for _, conn := range conns {
				err := conn.WriteJSON(data)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, userID := range userIDs {
		for _, conn := range n.ConnectionMap[userID] {
			err := conn.WriteJSON(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Control messages for websocket
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//return r.Header.Get("Origin") == "https://garson.me"
		return true
	},
}

//TODO: add a handler that upgrades clients to a WebSocket connection
//and adds that to a list of WebSockets to notify when events are
//read from the RabbitMQ server. Remember to synchronize changes
//to this list, as handlers are called concurrently from multiple
//goroutines.
func (ctx *HandlerCtx) WebSocketConnectionHandler(w http.ResponseWriter, r *http.Request) {
	sessionState := &SessionState{}
	if _, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, sessionState); err != nil {
		http.Error(w, "Unauthorized, log in first", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open websockect connection", 401)
		return
	}
	conn.WriteMessage(TextMessage, []byte("connection built"))

	userID := sessionState.AuthedUser.ID
	connID := ctx.Notifier.InsertConnection(userID, conn)

	go listen(conn, connID, userID, ctx.Notifier)
}

func listen(conn *websocket.Conn, connID int, userID int64, notifier *Notifier) {
	defer conn.Close()
	defer notifier.RemoveConnection(userID, connID)

	for {
		messageType, _, err := conn.ReadMessage()

		// if messageType == TextMessage || messageType == BinaryMessage {
		// 	notifier.WriteToConnections(TextMessage, append([]byte("Hello from server: "), p...))
		// } else
		if messageType == CloseMessage {
			println("Close message received for ws connection")
			break
		} else if err != nil {
			println("Error reading message")
			break
		}
		// ignore ping and pong messages
	}
}

type event struct {
	Type      string      `json:"type,omitempty"`
	Channel   interface{} `json:"channel,omitempty"`
	ChannelID string      `json:"channelID,omitempty"`
	UserIDs   []int64     `json:"userIDs,omitempty"`
	Message   interface{} `json:"message,omitempty"`
	MessageID string      `json:"messageID,omitempty"`
}

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections that should hear about
//that event. If you get an error writing to the WebSocket,
//just close it and remove it from the list
//(client went away without closing from
//their end). Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket
func (ctx *HandlerCtx) ConnectToRabbitAndListen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(
		"events", // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue")
	}

	msgs, err := ch.Consume(
		q.Name,       // queue
		"WebSockets", // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer")
	}

	go func() {
		for d := range msgs {
			e := &event{}
			err = json.Unmarshal(d.Body, e)
			if err != nil {
				log.Fatalf("Failed to read rabbit message (json marshal fail)")
			}

			err = ctx.Notifier.WriteToConnections(e, e.UserIDs)
			if err != nil {
				log.Fatalf("Failed to send messages from rabbit to ws")
			}
		}
	}()
}
