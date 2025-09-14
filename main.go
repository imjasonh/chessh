package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/gorilla/websocket"
	sshproxy "github.com/imjasonh/ssh-proxy"
)

type model struct {
	// Game state
	game       *Game
	cursorRow  int
	cursorCol  int
	selected   *Position
	validMoves []Position

	// Multiplayer state
	player      *Player
	opponent    *Player
	gameSession *GameSession
	gameState   string // "waiting", "playing", "finished", "opponent_disconnected"
	isMyTurn    bool
}

func initialModel() model {
	return model{
		game:       NewGame(),
		cursorRow:  0,
		cursorCol:  0,
		selected:   nil,
		validMoves: make([]Position, 0),
		gameState:  "waiting",
		isMyTurn:   false,
	}
}

func initialModelWithPlayer(player *Player) model {
	m := initialModel()
	m.player = player
	m.gameState = "waiting"
	return m
}

func (m model) Init() tea.Cmd {
	if m.player != nil && m.player.UpdateChan != nil {
		return m.listenForUpdates()
	}
	return nil
}

func (m model) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		if m.player != nil && m.player.UpdateChan != nil {
			return <-m.player.UpdateChan
		}
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEscape:
			if m.gameState == "playing" && m.isMyTurn {
				m.selected = nil
				m.validMoves = make([]Position, 0)
				// Broadcast deselection to opponent
				if m.gameSession != nil {
					GetGameManager().BroadcastUpdate(m.player.ID, GameUpdate{
						Type: "deselect",
						Data: nil,
					})
				}
			}
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		}

		// Only handle game input if it's the player's turn and game is active
		if m.gameState == "playing" && m.isMyTurn {
			switch msg.String() {
			case "up", "k":
				if m.cursorRow < 7 {
					m.cursorRow++
					m.broadcastCursorUpdate()
				}
			case "down", "j":
				if m.cursorRow > 0 {
					m.cursorRow--
					m.broadcastCursorUpdate()
				}
			case "left", "h":
				if m.cursorCol > 0 {
					m.cursorCol--
					m.broadcastCursorUpdate()
				}
			case "right", "l":
				if m.cursorCol < 7 {
					m.cursorCol++
					m.broadcastCursorUpdate()
				}
			case "enter", " ":
				currentPos := Position{m.cursorRow, m.cursorCol}

				if m.selected == nil {
					piece := m.game.Board.At(currentPos)
					if piece.Type != Empty && piece.Color == m.player.Color {
						m.selected = &currentPos
						m.validMoves = m.getValidMoves(currentPos)
						// Broadcast selection
						GetGameManager().BroadcastUpdate(m.player.ID, GameUpdate{
							Type: "select",
							Data: map[string]interface{}{
								"position":   currentPos,
								"validMoves": m.validMoves,
							},
						})
					}
				} else {
					if *m.selected == currentPos {
						m.selected = nil
						m.validMoves = make([]Position, 0)
						// Broadcast deselection
						GetGameManager().BroadcastUpdate(m.player.ID, GameUpdate{
							Type: "deselect",
							Data: nil,
						})
					} else if m.game.MakeMove(*m.selected, currentPos) {
						// Move successful - broadcast to opponent
						GetGameManager().BroadcastUpdate(m.player.ID, GameUpdate{
							Type: "move",
							Data: map[string]interface{}{
								"from":      *m.selected,
								"to":        currentPos,
								"gameState": m.game,
							},
						})
						m.selected = nil
						m.validMoves = make([]Position, 0)
						m.isMyTurn = false
					}
				}
			}
		} else if m.gameState == "waiting" || m.gameState == "opponent_disconnected" {
			// In waiting mode or after opponent disconnect, allow basic navigation for UI exploration but no moves
			switch msg.String() {
			case "up", "k":
				if m.cursorRow < 7 {
					m.cursorRow++
				}
			case "down", "j":
				if m.cursorRow > 0 {
					m.cursorRow--
				}
			case "left", "h":
				if m.cursorCol > 0 {
					m.cursorCol--
				}
			case "right", "l":
				if m.cursorCol < 7 {
					m.cursorCol++
				}
			}
		}

	case GameUpdate:
		return m.handleGameUpdate(msg)
	}
	return m, nil
}

func (m model) broadcastCursorUpdate() {
	if m.gameSession != nil {
		GetGameManager().BroadcastUpdate(m.player.ID, GameUpdate{
			Type: "cursor",
			Data: map[string]interface{}{
				"row": m.cursorRow,
				"col": m.cursorCol,
			},
		})
	}
}

func (m model) handleGameUpdate(update GameUpdate) (tea.Model, tea.Cmd) {
	// Don't process updates from self
	if m.player != nil && update.FromPlayer == m.player.ID {
		return m, m.listenForUpdates()
	}

	switch update.Type {
	case "matched":
		m.gameState = "playing"
		m.gameSession = GetGameManager().GetGameSession(m.player.ID)
		if m.gameSession != nil {
			m.game = m.gameSession.Game
			m.opponent = m.gameSession.GetOpponent(m.player.ID)
			m.isMyTurn = (m.player.Color == White) // White goes first
		}

	case "move":
		if data, ok := update.Data.(map[string]interface{}); ok {
			// Update game state from opponent's move
			if gameState, ok := data["gameState"].(*Game); ok {
				m.game = gameState
				m.isMyTurn = true // It's now our turn
			}
		}

	case "cursor":
		// Opponent cursor movement - we could show this in UI later

	case "select":
		// Opponent piece selection - we could show this in UI later

	case "deselect":
		// Opponent deselected - clear any opponent indicators

	case "opponent_disconnected":
		m.gameState = "opponent_disconnected"
		m.isMyTurn = false // Disable input
	}

	// Continue listening for updates
	return m, m.listenForUpdates()
}

func (m model) getValidMoves(from Position) []Position {
	var moves []Position

	for row := range 8 {
		for col := range 8 {
			to := Position{row, col}
			if m.game.IsValidMove(from, to) {
				moves = append(moves, to)
			}
		}
	}

	return moves
}

func (m model) View() string {
	var s strings.Builder

	if m.gameState == "waiting" {
		s.WriteString("CheSSH\n")
		s.WriteString("Waiting for an opponent to connect...\n\n")

		// Show queue position if available
		if m.player != nil {
			position := GetGameManager().GetQueuePosition(m.player.ID)
			if position > 0 {
				s.WriteString(fmt.Sprintf("Position in queue: %d\n", position))
			}
		}

		s.WriteString("You can explore the board while waiting:\n")
		s.WriteString("Use arrow keys to move cursor, Q to quit\n\n")
		s.WriteString(m.renderBoardWithInfo())
		return s.String()
	}

	if m.gameState == "opponent_disconnected" {
		s.WriteString("CheSSH\n")
		s.WriteString("*** OPPONENT DISCONNECTED; YOU WIN ***\n\n")
		s.WriteString("Your opponent has left the game.\n")
		s.WriteString("You can continue exploring the board or press Q to quit.\n\n")
		s.WriteString(m.renderBoardWithInfo())
		return s.String()
	}

	s.WriteString("CheSSH\n")

	if m.player != nil && m.opponent != nil {
		s.WriteString(fmt.Sprintf("You: %s (%s) vs %s (%s)\n",
			m.player.Name, m.player.Color, m.opponent.Name, m.opponent.Color))
	}

	if m.isMyTurn {
		s.WriteString("YOUR TURN - Use arrow keys to move cursor, ENTER/SPACE to select/move, ESC to deselect, Q to quit\n\n")
	} else {
		s.WriteString("OPPONENT'S TURN - Please wait for your opponent to move\n\n")
	}

	status := m.game.GameStatus()
	if status != "" {
		s.WriteString(fmt.Sprintf("*** %s ***\n\n", status))
	}

	s.WriteString(m.renderBoardWithInfo())

	return s.String()
}

func (m model) renderBoardWithInfo() string {
	boardLines := m.getBoardLines()
	infoLines := m.getInfoLines()

	var s strings.Builder
	maxLines := len(boardLines)
	if len(infoLines) > maxLines {
		maxLines = len(infoLines)
	}

	for i := 0; i < maxLines; i++ {
		if i < len(boardLines) {
			s.WriteString(boardLines[i])
		} else {
			s.WriteString(strings.Repeat(" ", 27)) // Board width padding
		}

		s.WriteString("   ")

		if i < len(infoLines) {
			s.WriteString(infoLines[i])
		}

		s.WriteString("\n")
	}

	return s.String()
}

func (m model) getBoardLines() []string {
	var lines []string

	lines = append(lines, "   a  b  c  d  e  f  g  h   ")

	for row := 7; row >= 0; row-- {
		var line strings.Builder
		line.WriteString(fmt.Sprintf("%d ", row+1))

		for col := 0; col < 8; col++ {
			pos := Position{row, col}
			piece := m.game.Board.At(pos)

			cellChar := piece.String()
			if piece.Type == Empty {
				if (row+col)%2 == 0 {
					cellChar = "·"
				} else {
					cellChar = " "
				}
			}

			if m.cursorRow == row && m.cursorCol == col {
				line.WriteString(fmt.Sprintf("[%s]", cellChar))
			} else if m.selected != nil && m.selected.Row == row && m.selected.Col == col {
				line.WriteString(fmt.Sprintf("<%s>", cellChar))
			} else if slices.Contains(m.validMoves, pos) {
				line.WriteString(fmt.Sprintf("*%s*", cellChar))
			} else {
				line.WriteString(fmt.Sprintf(" %s ", cellChar))
			}
		}

		line.WriteString(fmt.Sprintf(" %d", row+1))
		lines = append(lines, line.String())
	}

	lines = append(lines, "   a  b  c  d  e  f  g  h   ")

	return lines
}

func (m model) getInfoLines() []string {
	var lines []string

	lines = append(lines, "┌─────────────────────┐")
	lines = append(lines, "│ GAME INFO           │")
	lines = append(lines, "├─────────────────────┤")
	lines = append(lines, fmt.Sprintf("│ Turn: %-13s │", m.game.CurrentTurn))
	lines = append(lines, "│                     │")

	cursorPos := Position{m.cursorRow, m.cursorCol}
	piece := m.game.Board.At(cursorPos)
	lines = append(lines, fmt.Sprintf("│ Cursor: %-11s │", cursorPos.String()))

	if piece.Type == Empty {
		lines = append(lines, "│ Piece: Empty        │")
	} else {
		pieceName := m.getPieceName(piece)
		lines = append(lines, fmt.Sprintf("│ Piece: %-12s │", pieceName))
	}

	lines = append(lines, "│                     │")

	if m.selected != nil {
		lines = append(lines, "├─────────────────────┤")
		selectedPiece := m.game.Board.At(*m.selected)
		selectedName := m.getPieceName(selectedPiece)
		lines = append(lines, fmt.Sprintf("│ Selected: %-10s │", selectedName))
		lines = append(lines, fmt.Sprintf("│ At: %-15s │", m.selected.String()))

		if len(m.validMoves) > 0 {
			lines = append(lines, "│                     │")
			lines = append(lines, "│ Valid moves:        │")

			// Show up to 6 valid moves
			moveCount := len(m.validMoves)
			if moveCount > 6 {
				moveCount = 6
			}

			for i := 0; i < moveCount; i += 2 {
				var moveLine strings.Builder
				moveLine.WriteString("│ ")
				moveLine.WriteString(m.validMoves[i].String())

				if i+1 < moveCount {
					moveLine.WriteString(fmt.Sprintf("  %s", m.validMoves[i+1].String()))
				}

				// Pad to fit box width
				for moveLine.Len() < 20 {
					moveLine.WriteString(" ")
				}
				moveLine.WriteString(" │")

				lines = append(lines, moveLine.String())
			}

			if len(m.validMoves) > 6 {
				lines = append(lines, fmt.Sprintf("│ ... and %d more      │", len(m.validMoves)-6))
			}
		}
	}
	lines = append(lines, "└─────────────────────┘")

	if len(m.game.MoveHistory) > 0 {
		lastMove := m.game.MoveHistory[len(m.game.MoveHistory)-1]
		lines = append(lines, fmt.Sprintf("Last move: %s -> %-12s", lastMove.From.String(), lastMove.To.String()))
	}

	return lines
}

func (m model) getPieceName(piece Piece) string {
	if piece.Type == Empty {
		return "Empty"
	}

	color := "White"
	if piece.Color == Black {
		color = "Black"
	}

	pieceType := ""
	switch piece.Type {
	case Pawn:
		pieceType = "Pawn"
	case Rook:
		pieceType = "Rook"
	case Knight:
		pieceType = "Knight"
	case Bishop:
		pieceType = "Bishop"
	case Queen:
		pieceType = "Queen"
	case King:
		pieceType = "King"
	}

	return fmt.Sprintf("%s %s", color, pieceType)
}

func main() {
	var sshPort = flag.Int("port", 2222, "run as SSH server")
	flag.Parse()

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create Secret Manager client: %v", err)
	}
	defer client.Close()
	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: os.Getenv("SSH_HOST_KEY_SECRET"),
	})
	if err != nil {
		log.Fatalf("failed to access secret version: %v", err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf(":%d", *sshPort)),
		wish.WithHostKeyPEM(resp.Payload.Data),
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				// Create player from SSH session
				player := &Player{
					ID:         fmt.Sprintf("player_%d", time.Now().UnixNano()),
					Session:    s,
					Name:       s.User(),
					Connected:  true,
					UpdateChan: make(chan GameUpdate, 10),
				}

				// Create model with player
				m := initialModelWithPlayer(player)

				// Add player to matchmaking queue first
				GetGameManager().AddPlayer(player)

				// Handle cleanup on session end
				go func() {
					<-s.Context().Done()
					GetGameManager().RemovePlayer(player.ID)
					if player.UpdateChan != nil {
						close(player.UpdateChan)
					}
				}()

				return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithInput(s), tea.WithOutput(s)}
			}),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		log.Printf("Starting SSH chess server on :%d", *sshPort)
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	if httpPort := os.Getenv("PORT"); httpPort != "" {
		log.Print("Starting WebSocket to SSH proxy on port ", httpPort)
		http.HandleFunc("/ssh", sshproxy.ProxyWebSocketToSSH(fmt.Sprintf(":%d", *sshPort), websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin for now
			},
		}))
		if err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), nil); err != nil {
			log.Fatalln("HTTP server error:", err)
		}
	}

	<-ctx.Done()
	log.Println("Stopping SSH server")

	tctx, tcancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer tcancel()
	if err := s.Shutdown(tctx); err != nil {
		log.Fatalln(err)
	}
}
