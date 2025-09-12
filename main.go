package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	game       *Game
	cursorRow  int
	cursorCol  int
	selected   *Position
	validMoves []Position
}

func initialModel() model {
	return model{
		game:       NewGame(),
		cursorRow:  0,
		cursorCol:  0,
		selected:   nil,
		validMoves: make([]Position, 0),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEscape:
			m.selected = nil
			m.validMoves = make([]Position, 0)
		}
		
		switch msg.String() {
		case "q":
			return m, tea.Quit
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
		case "enter", " ":
			currentPos := Position{m.cursorRow, m.cursorCol}

			if m.selected == nil {
				piece := m.game.Board.At(currentPos)
				if piece.Type != Empty && piece.Color == m.game.CurrentTurn {
					m.selected = &currentPos
					m.validMoves = m.getValidMoves(currentPos)
				}
			} else {
				if *m.selected == currentPos {
					m.selected = nil
					m.validMoves = make([]Position, 0)
				} else if m.game.MakeMove(*m.selected, currentPos) {
					m.selected = nil
					m.validMoves = make([]Position, 0)
				}
			}
		}
	}
	return m, nil
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

	s.WriteString("Use arrow keys to move cursor, ENTER/SPACE to select/move, ESC to deselect, Q to quit\n\n")

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
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		log.Fatal(err)
	}
}
