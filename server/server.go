package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var username = ""

var GameStarted bool
var Timer20Started bool
var numOfPlayers int

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func NewClient(h *Hub, conn *websocket.Conn, username string) *Client {
	client := &Client{}
	if !GameStarted {
		client := &Client{
			Hub:           h,
			Conn:          conn,
			UserName:      username,
			UpdateChannel: make(chan interface{}),
		}
		h.Register <- client
		client.ID = len(h.Clients) + 1
		return client
	}
	return client
}

func (h *Hub) Run() {

	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

			h.registerPlayer()
			h.updatePlayerList()
			numOfPlayers = len(h.Clients)

			h.timer(&numOfPlayers, &Timer20Started)
		case client := <-h.Unregister:

			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				h.updatePlayerList()
			}
		}
	}
}

func (h *Hub) timer(numOfPlayers *int, Timer20Started *bool) {

	if !*Timer20Started && *numOfPlayers > 1 {
		ticker := time.NewTicker(1 * time.Second)

		*Timer20Started = true
		countdown := 20

		go func() {
			for countdown > 0 {
				select {
				case <-ticker.C:
					log.Println("20-second Timer:", countdown)
					countdown--
				}
				if *numOfPlayers == 4 {
					h.start10SecTimer()
					ticker.Stop()
				}
			}

			if countdown == 0 {
				h.start10SecTimer()
				ticker.Stop()
			}
		}()
	}
}

func (h *Hub) start10SecTimer() {
	log.Println("Starting 10-second countdown...")
	ticker := time.NewTicker(1 * time.Second)
	countdown := 10
	GameStarted = true

	go func() {
		for countdown > 0 {
			select {
			case <-ticker.C:
				log.Println("10-second Timer:", countdown)
				countdown--
				message := SocketMessage{
					Type:  "timer",
					Timer: countdown,
				}

				h.SendUpdate(message)
			}
		}

		if countdown == 0 {
			ticker.Stop()
			h.startGame()
		}
	}()
}

func (h *Hub) updatePlayerList() {
	usernames := make([]string, 0, len(h.Clients))
	for client := range h.Clients {
		usernames = append(usernames, client.UserName)
	}

	message := SocketMessage{
		Type:      "player_list",
		Usernames: usernames,
	}

	h.SendUpdate(message)
}

func (h *Hub) SendUpdate(data interface{}) {

	for client := range h.Clients {
		err := client.Conn.WriteJSON(data)
		if err != nil {
			log.Println("Error writing JSON:", err)
		}
	}
}

func SendWSMessage(conn *websocket.Conn, message SocketMessage) {
	msgBytes, _ := json.Marshal(message)
	conn.WriteMessage(websocket.TextMessage, msgBytes)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request, h *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	username = r.URL.Query().Get("username")

	if len(h.Clients) == 4 {
		response := SocketMessage{
			Type: "server_full"}
		SendWSMessage(conn, response)
		return
	}

	for client := range h.Clients {
		if client.UserName == username {

			response := SocketMessage{
				Type: "user_taken"}
			SendWSMessage(conn, response)
			return
		}
	}

	client := NewClient(h, conn, username)

	defer func() {
		h.Unregister <- client
		conn.Close()
	}()

	for {
		var data json.RawMessage
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Println(err)
			break
		}

		var message SocketMessage
		if err := json.Unmarshal(data, &message); err != nil {
			log.Println("Error unmarshal message:", err)
			continue
		}

		switch message.Type {
		case "chat":
			var chatMessage ChatMessage
			if err := json.Unmarshal(data, &chatMessage); err != nil {
				log.Println("Error unmarshal chat message:", err)
				continue
			}
			h.SendUpdate(chatMessage)

		case "keypress":
			var keyMessage keyMessage
			if err := json.Unmarshal(data, &keyMessage); err != nil {
				log.Println("Error unmarshal keypress message:", err)
				continue
			}

			h.handleKeyPress(keyMessage.Key, keyMessage.UserName)
		}
	}
}

func (h *Hub) registerPlayer() {

	if !GameStarted {
		response := SocketMessage{
			Type:   "user_added",
			Sender: username,
		}

		h.SendUpdate(response)
		username = ""
	}
}
