package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/ssh"
)

// Player represents a connected player
type Player struct {
	ID         string
	Session    ssh.Session
	Color      Color
	Name       string
	GameID     string
	Connected  bool
	UpdateChan chan GameUpdate // Channel for sending updates to the player's model
}

// GameUpdate represents an update to broadcast to players
type GameUpdate struct {
	Type       string      // "move", "cursor", "select", "gamestate"
	Data       interface{} // The actual update data
	FromPlayer string      // Which player sent the update
}

// GameSession manages a single game between two players
type GameSession struct {
	ID      string
	Game    *Game
	White   *Player
	Black   *Player
	Updates chan GameUpdate
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
}

func NewGameSession(id string, white, black *Player) *GameSession {
	ctx, cancel := context.WithCancel(context.Background())

	session := &GameSession{
		ID:      id,
		Game:    NewGame(),
		White:   white,
		Black:   black,
		Updates: make(chan GameUpdate, 10),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Assign colors and game ID to players
	white.Color = White
	white.GameID = id
	black.Color = Black
	black.GameID = id

	// Start the update broadcaster
	go session.handleUpdates()

	return session
}

func (gs *GameSession) handleUpdates() {
	for {
		select {
		case <-gs.ctx.Done():
			return
		case update := <-gs.Updates:
			gs.broadcastUpdate(update)
		}
	}
}

func (gs *GameSession) broadcastUpdate(update GameUpdate) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Send update to both players (if connected) via their update channels
	if gs.White != nil && gs.White.Connected && gs.White.UpdateChan != nil {
		select {
		case gs.White.UpdateChan <- update:
		default:
			// Channel full, drop update
		}
	}
	if gs.Black != nil && gs.Black.Connected && gs.Black.UpdateChan != nil {
		select {
		case gs.Black.UpdateChan <- update:
		default:
			// Channel full, drop update
		}
	}
}

func (gs *GameSession) GetPlayer(playerID string) *Player {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.White != nil && gs.White.ID == playerID {
		return gs.White
	}
	if gs.Black != nil && gs.Black.ID == playerID {
		return gs.Black
	}
	return nil
}

func (gs *GameSession) GetOpponent(playerID string) *Player {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if gs.White != nil && gs.White.ID == playerID {
		return gs.Black
	}
	if gs.Black != nil && gs.Black.ID == playerID {
		return gs.White
	}
	return nil
}

func (gs *GameSession) IsPlayerTurn(playerID string) bool {
	player := gs.GetPlayer(playerID)
	if player == nil {
		return false
	}
	return gs.Game.CurrentTurn == player.Color
}

func (gs *GameSession) Disconnect(playerID string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	var disconnectedPlayer, remainingPlayer *Player

	if gs.White != nil && gs.White.ID == playerID {
		gs.White.Connected = false
		disconnectedPlayer = gs.White
		remainingPlayer = gs.Black
	}
	if gs.Black != nil && gs.Black.ID == playerID {
		gs.Black.Connected = false
		disconnectedPlayer = gs.Black
		remainingPlayer = gs.White
	}

	// Notify remaining player of opponent disconnect
	if remainingPlayer != nil && remainingPlayer.Connected && remainingPlayer.UpdateChan != nil {
		disconnectUpdate := GameUpdate{
			Type: "opponent_disconnected",
			Data: map[string]any{
				"disconnectedPlayer": disconnectedPlayer.Name,
			},
		}
		select {
		case remainingPlayer.UpdateChan <- disconnectUpdate:
		default:
		}
	}

	// If both players disconnected, cleanup
	if (gs.White == nil || !gs.White.Connected) && (gs.Black == nil || !gs.Black.Connected) {
		gs.cleanup()
	}
}

func (gs *GameSession) cleanup() {
	gs.cancel()
	close(gs.Updates)
}

// GameManager handles matchmaking and game coordination
type GameManager struct {
	playerQueue  []*Player
	activeGames  map[string]*GameSession
	playerToGame map[string]string // playerID -> gameID
	mu           sync.RWMutex
	gameCounter  int
}

var gameManager *GameManager
var gameManagerOnce sync.Once

func GetGameManager() *GameManager {
	gameManagerOnce.Do(func() {
		gameManager = &GameManager{
			playerQueue:  make([]*Player, 0),
			activeGames:  make(map[string]*GameSession),
			playerToGame: make(map[string]string),
		}
	})
	return gameManager
}

func (gm *GameManager) AddPlayer(player *Player) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Add to queue
	gm.playerQueue = append(gm.playerQueue, player)

	// Try to match with another player
	if len(gm.playerQueue) >= 2 {
		white := gm.playerQueue[0]
		black := gm.playerQueue[1]

		// Remove from queue
		gm.playerQueue = gm.playerQueue[2:]

		// Create game
		gm.gameCounter++
		gameID := fmt.Sprintf("game_%d", gm.gameCounter)

		session := NewGameSession(gameID, white, black)
		gm.activeGames[gameID] = session
		gm.playerToGame[white.ID] = gameID
		gm.playerToGame[black.ID] = gameID

		// Notify players they've been matched
		matchUpdate := GameUpdate{
			Type: "matched",
			Data: map[string]any{
				"gameID": gameID,
				"opponent": map[string]string{
					"white_opponent": black.Name,
					"black_opponent": white.Name,
				},
			},
		}

		// Send match update to both players via their channels
		if white.UpdateChan != nil {
			select {
			case white.UpdateChan <- matchUpdate:
			default:
			}
		}
		if black.UpdateChan != nil {
			select {
			case black.UpdateChan <- matchUpdate:
			default:
			}
		}
	}
}

func (gm *GameManager) RemovePlayer(playerID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Remove from queue if present
	for i, player := range gm.playerQueue {
		if player.ID == playerID {
			gm.playerQueue = append(gm.playerQueue[:i], gm.playerQueue[i+1:]...)
			break
		}
	}

	// Handle active game disconnection
	if gameID, exists := gm.playerToGame[playerID]; exists {
		if session, gameExists := gm.activeGames[gameID]; gameExists {
			session.Disconnect(playerID)

			// Clean up if game is over
			if (session.White == nil || !session.White.Connected) &&
				(session.Black == nil || !session.Black.Connected) {
				delete(gm.activeGames, gameID)
				delete(gm.playerToGame, session.White.ID)
				delete(gm.playerToGame, session.Black.ID)
			}
		}
		delete(gm.playerToGame, playerID)
	}
}

func (gm *GameManager) GetGameSession(playerID string) *GameSession {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	if gameID, exists := gm.playerToGame[playerID]; exists {
		return gm.activeGames[gameID]
	}
	return nil
}

func (gm *GameManager) BroadcastUpdate(playerID string, update GameUpdate) {
	session := gm.GetGameSession(playerID)
	if session != nil {
		update.FromPlayer = playerID
		select {
		case session.Updates <- update:
		case <-time.After(100 * time.Millisecond):
			// Drop update if channel is full
		}
	}
}

func (gm *GameManager) GetQueuePosition(playerID string) int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for i, player := range gm.playerQueue {
		if player.ID == playerID {
			return i + 1
		}
	}
	return -1
}
