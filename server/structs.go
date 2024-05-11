package server

import (
	"time"

	"github.com/gorilla/websocket"
)

// Hub represents a hub for managing client connections and message broadcasting
type Hub struct {
	GameState  GameState
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
}

// Client represents a client connected to the server
type Client struct {
	ID            int
	Hub           *Hub
	Conn          *websocket.Conn
	UserName      string
	UpdateChannel chan interface{}
}

type SocketMessage struct {
	Type         string   `json:"type"`
	Sender       string   `json:"sender,omitempty"`
	Usernames    []string `json:"usernames,omitempty"`
	Timer        int      `json:"timer,omitempty"`
	TimerStarted bool     `json:"timer_started,omitempty"`
}

type ChatMessage struct {
	Type      string    `json:"type"`
	Sender    string    `json:"sender"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type keyMessage struct {
	Type     string `json:"type"`
	Key      string `json:"key"`
	UserName string `json:"user_name"`
}

type GameState struct {
	Type    string     `json:"type"`
	Board   [][]string `json:"board"`
	Players []Player   `json:"players"`
	Bombs   []Bomb     `json:"bombs"`
}

type Player struct {
	ID             string    `json:"id"`
	UserName       string    `json:"user_name"`
	Lives          int       `json:"lives"`
	Position       Position  `json:"position"`
	BombCapacity   int       `json:"bomb_capacity"`
	ExplosionRange int       `json:"explosion_range"`
	Speed          float64   `json:"speed"`
	IsAlive        bool      `json:"is_alive"`
	LastMoveTime   time.Time `json:"last_move,omitempty"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Bomb struct {
	Position       Position `json:"position"`
	ExplosionRange int      `json:"explosion_range"`
}
