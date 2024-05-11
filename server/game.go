package server

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

const (
	cols          int    = 15
	rows          int    = 13
	empty         string = " "
	wall          string = "X"
	block         string = "B"
	blastRange    string = "#"
	blockChance   int    = 66
	powerUpChance int    = 30
)

var readyTomove bool

var powerUps = []string{"power-up-bombs", "power-up-flames", "power-up-speed"}

func GenerateGameBoard() [][]string {
	board := make([][]string, rows)

	// Create the outer borders
	for i := range board {
		board[i] = make([]string, cols)

		for j := range board[i] {
			if i == 0 || i == rows-1 || j == 0 || j == cols-1 {
				board[i][j] = wall
			} else if i%2 == 0 && j%2 == 0 {
				board[i][j] = wall
			} else {
				// Initialize the inner area with empty tiles
				board[i][j] = empty
			}
		}
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Skip the cells adjacent to the corners
			if (i >= 1 && i <= 2 && j >= 1 && j <= 2) ||
				(i >= 1 && i <= 2 && j >= cols-3 && j <= cols-2) ||
				(i >= rows-3 && i <= rows-2 && j >= 1 && j <= 2) ||
				(i >= rows-3 && i <= rows-2 && j >= cols-3 && j <= cols-2) {
				continue
			}

			// 66% of possibility to place a block block
			if board[i][j] == empty {
				randomNumber := rand.Intn(100)
				if randomNumber < blockChance {
					board[i][j] = block
				}
			}
		}
	}

	return board
}

func (h *Hub) setPlayers() (players []Player) {
	// Define starting positions for players
	startingPositions := []Position{
		{1, 1},
		{11, 13},
		{1, 13},
		{11, 1},
	}

	// Initialize a counter to keep track of the number of players
	countPlayers := 0

	// Assign players to their starting positions and populate the players slice
	for client := range h.Clients {
		if countPlayers >= len(startingPositions) {
			break
		}
		player := Player{
			ID:             strconv.Itoa(countPlayers + 1),
			UserName:       client.UserName,
			Position:       startingPositions[countPlayers],
			Lives:          3,
			BombCapacity:   1,
			ExplosionRange: 1,
			Speed:          1,
			IsAlive:        true,
		}
		players = append(players, player)
		countPlayers++
	}

	return players
}

// Function to start the game
func (h *Hub) startGame() {
	log.Println("Game started!")

	h.GameState.Type = "loadBoard"
	h.GameState.Board = GenerateGameBoard()
	h.GameState.Players = h.setPlayers()

	h.SendUpdate(h.GameState)
}

func (h *Hub) handleKeyPress(key string, username string) {
	// Find the index of the player in the slice
	var playerIndex int
	found := false
	for i, p := range h.GameState.Players {
		if p.UserName == username {
			playerIndex = i
			found = true
			break
		}
	}

	if !found {
		log.Println("Player is dead:", username)
		return
	}

	// Reference the actual player from the slice
	player := &h.GameState.Players[playerIndex]

	// Calculate cooldown duration based on player's speed
	cooldownDuration := time.Duration(200/player.Speed) * time.Millisecond

	// Define a function to check movement cooldown
	movingCoolDown := func() bool {
		if time.Now().Sub(player.LastMoveTime) < cooldownDuration {
			return false
		}
		return true
	}

	// Perform actions based on the pressed key
	switch key {
	case "ArrowLeft":
		if movingCoolDown() && player.Position.Y > 1 && h.isValidMove(player.Position.X, player.Position.Y-1, h.GameState.Board, player) {
			player.Position.Y--
			player.LastMoveTime = time.Now() // Update last move time
		}
	case "ArrowRight":
		if movingCoolDown() && player.Position.Y < 13 && h.isValidMove(player.Position.X, player.Position.Y+1, h.GameState.Board, player) {
			player.Position.Y++
			player.LastMoveTime = time.Now() // Update last move time
		}
	case "ArrowUp":
		if movingCoolDown() && player.Position.X > 1 && h.isValidMove(player.Position.X-1, player.Position.Y, h.GameState.Board, player) {
			player.Position.X--
			player.LastMoveTime = time.Now() // Update last move time
		}
	case "ArrowDown":
		if movingCoolDown() && player.Position.X < 11 && h.isValidMove(player.Position.X+1, player.Position.Y, h.GameState.Board, player) {
			player.Position.X++
			player.LastMoveTime = time.Now() // Update last move time
		}
	case " ":
		if player.BombCapacity > 0 {
			h.PlaceBomb(player)
			h.GameState.Type = "bomb"
			h.SendUpdate(h.GameState)
		}
	}

	h.GameState.Type = "playerAction"
	h.SendUpdate(h.GameState)
}

// Function to check if the move is valid
func (h *Hub) isValidMove(x, y int, board [][]string, player *Player) bool {
	if board[x][y] == "power-up-bombs" {
		// change bombCapacity of that player
		player.BombCapacity++
		h.GameState.Board[x][y] = " "

		return true
	}
	if board[x][y] == "power-up-flames" {
		// change explosionRange of that player
		player.ExplosionRange++
		h.GameState.Board[x][y] = " "

		return true
	}
	if board[x][y] == "power-up-speed" {
		// change speed of that player
		player.Speed++
		h.GameState.Board[x][y] = " "

		return true
	}

	return board[x][y] == " "
}

// Function to place bomb
func (h *Hub) PlaceBomb(player *Player) {

	bomb := Bomb{
		Position:       player.Position,
		ExplosionRange: player.ExplosionRange,
	}
	h.GameState.Bombs = append(h.GameState.Bombs, bomb)
	player.BombCapacity-- // Decrement bomb capacity after placing bomb

	h.GameState.Board[bomb.Position.X][bomb.Position.Y] = "bomb"

	// Schedule the explosion after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		h.ExplodeBomb(bomb)

		player.BombCapacity++ // increment bomb capacity again after placing bomb
		for i, b := range h.GameState.Bombs {
			if b == bomb {
				h.GameState.Bombs = append(h.GameState.Bombs[:i], h.GameState.Bombs[i+1:]...)
				break
			}
		}
	})
}

// Function to handle bomb explosion
func (h *Hub) ExplodeBomb(bomb Bomb) {
	// Calculate the affected tiles within the explosion range
	affectedTiles := calculateAffectedTiles(bomb, h.GameState.Board)
	blockDestroyedTiles := []Position{} // List to store tiles where blocks were destroyed

	// Handle destruction of blocks and elimination of players
	for _, tile := range affectedTiles {
		tileContent := h.GameState.Board[tile.X][tile.Y]
		if player := h.GetPlayerAt(tile); player != -1 {
			// Eliminate players within the explosion range
			h.EliminatePlayer(player)
		}

		if tileContent == "B" || tileContent == "bomb" || tileContent == " " || tileContent == "power-up-bombs" || tileContent == "power-up-flames" || tileContent == "power-up-speed" {
			if tileContent == "B" {
				blockDestroyedTiles = append(blockDestroyedTiles, tile) // Save the tile
			}

			h.GameState.Board[tile.X][tile.Y] = "#"
		}
	}

	// Send the updated game state to all clients after the explosion
	h.GameState.Type = "explosion"
	h.SendUpdate(h.GameState)

	time.AfterFunc(500*time.Millisecond, func() { // Handle destruction of blocks and elimination of players
		for _, tile := range affectedTiles {
			tileContent := h.GameState.Board[tile.X][tile.Y]
			if tileContent == "#" || tileContent == " " {
				// Schedule the removal of destroyed blocks after 0.5 seconds
				h.GameState.Board[tile.X][tile.Y] = " "
			}
		}

		for _, tile := range blockDestroyedTiles {
			// Check if a power-up should be placed
			if rand.Intn(100) < powerUpChance { // 30% chance for power-up
				// Choose a random power-up from the slice
				powerUpIndex := rand.Intn(len(powerUps))
				h.GameState.Board[tile.X][tile.Y] = powerUps[powerUpIndex]
			}
		}

		h.SendUpdate(h.GameState)
	})
}

func (h *Hub) GetPlayerAt(position Position) int {
	for i, player := range h.GameState.Players {
		if player.Position == position {
			return i
		}
	}
	return -1
}

// Function to calculate the affected tiles within the explosion range
func calculateAffectedTiles(bomb Bomb, board [][]string) []Position {
	var affectedTiles []Position
	bombPos := make(map[Position]bool)

	// Add the bomb's position as the central tile
	affectedTiles = append(affectedTiles, bomb.Position)
	bombPos[bomb.Position] = true

	// Directions to explore: left, right, up, down
	directions := [][]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	// Explore in each direction within the explosion range
	for _, dir := range directions {
		exploreDirection(bomb.Position.X, bomb.Position.Y, dir[0], dir[1], bomb.ExplosionRange, board, &affectedTiles, bombPos)
	}

	return affectedTiles
}

// Function to explore tiles in a specific direction within the explosion range
func exploreDirection(x, y, dx, dy, explosionRange int, board [][]string, affectedTiles *[]Position, bombPos map[Position]bool) {
	// Base case: Stop exploring if the explosion range is exhausted or if an indestructible wall is encountered
	for explosionRange > 0 {
		newX, newY := x+dx, y+dy
		newPos := Position{X: newX, Y: newY}

		// Check if the new position is within the bounds of the board and not bombPos
		if newX >= 0 && newX < len(board) && newY >= 0 && newY < len(board[0]) && !bombPos[newPos] {
			bombPos[newPos] = true

			// Check if the tile is destructible
			if isDestructible(newX, newY, board) {
				if board[x][y] == block {
					break
				}
				// Add the destructible tile to the affected tiles list
				*affectedTiles = append(*affectedTiles, newPos)
				// Move to the next tile in the same direction
				x, y = newX, newY
				explosionRange--

			} else {
				// Stop exploring in this direction if an indestructible wall is encountered
				break
			}
		} else {
			// Stop exploring if the new position is out of bounds
			break
		}
	}
}

// Function to check if a tile is destructible
func isDestructible(x, y int, board [][]string) bool {
	// Check if the tile is within the bounds of the board
	if x < 0 || x >= len(board) || y < 0 || y >= len(board[0]) {
		return false
	}

	// Check if the tile is destructible (not a wall)
	return board[x][y] != wall
}

// Function to eliminate a player
func (h *Hub) EliminatePlayer(playerIndex int) {
	player := &h.GameState.Players[playerIndex]
	player.Lives--
	log.Println("Player: "+player.UserName+" lost a life, lives left:", player.Lives)

	// If the player has no lives left, remove them from the game
	if player.Lives <= 0 {
		// Remove the player from the slice
		h.GameState.Players = append(h.GameState.Players[:playerIndex], h.GameState.Players[playerIndex+1:]...)
		h.SendUpdate(h.GameState)
		// Alert GAME OVER if needed
		if len(h.GameState.Players) == 0 {
			// If there are no more players left, alert GAME OVER
			log.Println("GAME OVER")
		}
	}
}
