package main

import "fmt"

type Color int

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	if c == White {
		return "White"
	}
	return "Black"
}

type PieceType int

const (
	Empty PieceType = iota
	Pawn
	Rook
	Knight
	Bishop
	Queen
	King
)

type Piece struct {
	Type  PieceType
	Color Color
}

func (p Piece) String() string {
	if p.Type == Empty {
		return " "
	}

	symbols := map[PieceType]string{
		Pawn:   "♟",
		Rook:   "♜",
		Knight: "♞",
		Bishop: "♝",
		Queen:  "♛",
		King:   "♚",
	}

	if p.Color == White {
		whiteSymbols := map[PieceType]string{
			Pawn:   "♙",
			Rook:   "♖",
			Knight: "♘",
			Bishop: "♗",
			Queen:  "♕",
			King:   "♔",
		}
		return whiteSymbols[p.Type]
	}

	return symbols[p.Type]
}

type Position struct {
	Row, Col int
}

func (p Position) Valid() bool {
	return p.Row >= 0 && p.Row < 8 && p.Col >= 0 && p.Col < 8
}

func (p Position) String() string {
	if !p.Valid() {
		return "invalid"
	}
	return fmt.Sprintf("%c%d", 'a'+p.Col, p.Row+1)
}

type Move struct {
	From, To Position
	Piece    Piece
}

type Board [8][8]Piece

func NewBoard() Board {
	var board Board

	board[0] = [8]Piece{
		{Rook, White}, {Knight, White}, {Bishop, White}, {Queen, White},
		{King, White}, {Bishop, White}, {Knight, White}, {Rook, White},
	}
	board[1] = [8]Piece{
		{Pawn, White}, {Pawn, White}, {Pawn, White}, {Pawn, White},
		{Pawn, White}, {Pawn, White}, {Pawn, White}, {Pawn, White},
	}

	for col := range 8 {
		for row := 2; row < 6; row++ {
			board[row][col] = Piece{Empty, White}
		}
	}

	board[6] = [8]Piece{
		{Pawn, Black}, {Pawn, Black}, {Pawn, Black}, {Pawn, Black},
		{Pawn, Black}, {Pawn, Black}, {Pawn, Black}, {Pawn, Black},
	}
	board[7] = [8]Piece{
		{Rook, Black}, {Knight, Black}, {Bishop, Black}, {Queen, Black},
		{King, Black}, {Bishop, Black}, {Knight, Black}, {Rook, Black},
	}

	return board
}

func (b *Board) At(pos Position) Piece {
	if !pos.Valid() {
		return Piece{Empty, White}
	}
	return b[pos.Row][pos.Col]
}

func (b *Board) Set(pos Position, piece Piece) {
	if pos.Valid() {
		b[pos.Row][pos.Col] = piece
	}
}

func (b *Board) Move(from, to Position) bool {
	if !from.Valid() || !to.Valid() {
		return false
	}

	piece := b.At(from)
	if piece.Type == Empty {
		return false
	}

	b.Set(to, piece)
	b.Set(from, Piece{Empty, White})
	return true
}

type Game struct {
	Board           Board
	CurrentTurn     Color
	MoveHistory     []Move
	KingMoved       [2]bool
	RookMoved       [2][2]bool
	EnPassantTarget *Position
}

func NewGame() *Game {
	return &Game{
		Board:           NewBoard(),
		CurrentTurn:     White,
		MoveHistory:     make([]Move, 0),
		KingMoved:       [2]bool{false, false},
		RookMoved:       [2][2]bool{{false, false}, {false, false}},
		EnPassantTarget: nil,
	}
}

func (g *Game) IsValidMove(from, to Position) bool {
	if !from.Valid() || !to.Valid() {
		return false
	}

	piece := g.Board.At(from)
	if piece.Type == Empty || piece.Color != g.CurrentTurn {
		return false
	}

	target := g.Board.At(to)
	if target.Type != Empty && target.Color == piece.Color {
		return false
	}

	return g.isValidPieceMove(piece, from, to)
}

func (g *Game) isValidPieceMove(piece Piece, from, to Position) bool {
	dx := to.Col - from.Col
	dy := to.Row - from.Row

	switch piece.Type {
	case Pawn:
		return g.isValidPawnMove(piece, from, to, dx, dy)
	case Rook:
		return g.isValidRookMove(from, to, dx, dy)
	case Knight:
		return g.isValidKnightMove(dx, dy)
	case Bishop:
		return g.isValidBishopMove(from, to, dx, dy)
	case Queen:
		return g.isValidQueenMove(from, to, dx, dy)
	case King:
		return g.isValidKingMove(piece, dx, dy)
	}
	return false
}

func (g *Game) isValidPawnMove(piece Piece, from, to Position, dx, dy int) bool {
	direction := 1
	startRow := 1
	if piece.Color == Black {
		direction = -1
		startRow = 6
	}

	if dx == 0 && g.Board.At(to).Type == Empty {
		if dy == direction {
			return true
		}
		if from.Row == startRow && dy == 2*direction {
			return true
		}
	}

	if abs(dx) == 1 && dy == direction {
		target := g.Board.At(to)
		if target.Type != Empty && target.Color != piece.Color {
			return true
		}

		if g.EnPassantTarget != nil && *g.EnPassantTarget == to {
			return true
		}
	}

	return false
}

func (g *Game) isValidRookMove(from, to Position, dx, dy int) bool {
	if dx != 0 && dy != 0 {
		return false
	}
	return g.isPathClear(from, to)
}

func (g *Game) isValidKnightMove(dx, dy int) bool {
	return (abs(dx) == 2 && abs(dy) == 1) || (abs(dx) == 1 && abs(dy) == 2)
}

func (g *Game) isValidBishopMove(from, to Position, dx, dy int) bool {
	if abs(dx) != abs(dy) {
		return false
	}
	return g.isPathClear(from, to)
}

func (g *Game) isValidQueenMove(from, to Position, dx, dy int) bool {
	return g.isValidRookMove(from, to, dx, dy) || g.isValidBishopMove(from, to, dx, dy)
}

func (g *Game) isValidKingMove(piece Piece, dx, dy int) bool {
	if abs(dx) <= 1 && abs(dy) <= 1 {
		return true
	}

	if abs(dx) == 2 && dy == 0 && !g.KingMoved[piece.Color] {
		return g.canCastle(piece.Color, dx > 0)
	}

	return false
}

func (g *Game) canCastle(color Color, kingSide bool) bool {
	row := 0
	if color == Black {
		row = 7
	}

	rookCol := 0
	if kingSide {
		rookCol = 7
	}

	if g.RookMoved[color][rookCol/7] {
		return false
	}

	if g.IsInCheck(color) {
		return false
	}

	start, end := 1, 3
	if kingSide {
		start, end = 5, 6
	}

	for col := start; col <= end; col++ {
		if g.Board.At(Position{row, col}).Type != Empty {
			return false
		}

		g.Board.Set(Position{row, 4}, Piece{Empty, White})
		g.Board.Set(Position{row, col}, Piece{King, color})
		inCheck := g.IsInCheck(color)
		g.Board.Set(Position{row, 4}, Piece{King, color})
		g.Board.Set(Position{row, col}, Piece{Empty, White})

		if inCheck {
			return false
		}
	}

	return true
}

func (g *Game) isPathClear(from, to Position) bool {
	dx := sign(to.Col - from.Col)
	dy := sign(to.Row - from.Row)

	current := Position{from.Row + dy, from.Col + dx}

	for current != to {
		if g.Board.At(current).Type != Empty {
			return false
		}
		current.Row += dy
		current.Col += dx
	}

	return true
}

func (g *Game) MakeMove(from, to Position) bool {
	if !g.IsValidMove(from, to) {
		return false
	}

	piece := g.Board.At(from)
	move := Move{from, to, piece}
	originalTarget := g.Board.At(to)

	g.executeMove(from, to, piece)

	if g.IsInCheck(g.CurrentTurn) {
		g.undoMove(from, to, piece, originalTarget)
		return false
	}

	g.updateGameState(from, to, piece)
	g.MoveHistory = append(g.MoveHistory, move)
	g.CurrentTurn = 1 - g.CurrentTurn

	return true
}

func (g *Game) executeMove(from, to Position, piece Piece) {
	if piece.Type == King && abs(to.Col-from.Col) == 2 {
		g.executeCastle(from, to)
	} else if piece.Type == Pawn && g.EnPassantTarget != nil && *g.EnPassantTarget == to {
		g.executeEnPassant(from, to)
	} else {
		g.Board.Move(from, to)
	}
}

func (g *Game) executeCastle(from, to Position) {
	g.Board.Move(from, to)

	rookFromCol := 0
	rookToCol := 3
	if to.Col > from.Col {
		rookFromCol = 7
		rookToCol = 5
	}

	rookFrom := Position{from.Row, rookFromCol}
	rookTo := Position{from.Row, rookToCol}
	g.Board.Move(rookFrom, rookTo)
}

func (g *Game) executeEnPassant(from, to Position) {
	g.Board.Move(from, to)
	capturedPawnRow := from.Row
	g.Board.Set(Position{capturedPawnRow, to.Col}, Piece{Empty, White})
}

func (g *Game) undoMove(from, to Position, piece Piece, originalTarget Piece) {
	if piece.Type == King && abs(to.Col-from.Col) == 2 {
		g.undoCastle(from, to, piece.Color)
	} else if piece.Type == Pawn && g.EnPassantTarget != nil && *g.EnPassantTarget == to {
		g.undoEnPassant(from, to)
	} else {
		g.Board.Set(from, piece)
		g.Board.Set(to, originalTarget)
	}
}

func (g *Game) undoCastle(from, to Position, color Color) {
	g.Board.Set(from, Piece{King, color})
	g.Board.Set(to, Piece{Empty, White})

	rookFromCol := 0
	rookToCol := 3
	if to.Col > from.Col {
		rookFromCol = 7
		rookToCol = 5
	}

	rookFrom := Position{from.Row, rookFromCol}
	rookTo := Position{from.Row, rookToCol}
	g.Board.Set(rookFrom, Piece{Rook, color})
	g.Board.Set(rookTo, Piece{Empty, White})
}

func (g *Game) undoEnPassant(from, to Position) {
	g.Board.Set(from, Piece{Pawn, g.CurrentTurn})
	g.Board.Set(to, Piece{Empty, White})
	capturedPawnRow := from.Row
	enemyColor := 1 - g.CurrentTurn
	g.Board.Set(Position{capturedPawnRow, to.Col}, Piece{Pawn, enemyColor})
}

func (g *Game) updateGameState(from, to Position, piece Piece) {
	g.EnPassantTarget = nil

	if piece.Type == King {
		g.KingMoved[piece.Color] = true
	} else if piece.Type == Rook {
		switch from.Col {
		case 0:
			g.RookMoved[piece.Color][0] = true
		case 7:
			g.RookMoved[piece.Color][1] = true
		}
	} else if piece.Type == Pawn && abs(to.Row-from.Row) == 2 {
		enPassantRow := (from.Row + to.Row) / 2
		g.EnPassantTarget = &Position{enPassantRow, from.Col}
	}
}

func (g *Game) FindKing(color Color) Position {
	for row := range 8 {
		for col := range 8 {
			pos := Position{row, col}
			piece := g.Board.At(pos)
			if piece.Type == King && piece.Color == color {
				return pos
			}
		}
	}
	return Position{-1, -1}
}

func (g *Game) IsInCheck(color Color) bool {
	kingPos := g.FindKing(color)
	if !kingPos.Valid() {
		return false
	}

	enemyColor := 1 - color

	for row := range 8 {
		for col := range 8 {
			pos := Position{row, col}
			piece := g.Board.At(pos)
			if piece.Type != Empty && piece.Color == enemyColor {
				if g.canPieceAttack(piece, pos, kingPos) {
					return true
				}
			}
		}
	}

	return false
}

func (g *Game) canPieceAttack(piece Piece, from, to Position) bool {
	dx := to.Col - from.Col
	dy := to.Row - from.Row

	switch piece.Type {
	case Pawn:
		direction := 1
		if piece.Color == Black {
			direction = -1
		}
		return abs(dx) == 1 && dy == direction
	case Rook:
		return (dx == 0 || dy == 0) && g.isPathClear(from, to)
	case Knight:
		return (abs(dx) == 2 && abs(dy) == 1) || (abs(dx) == 1 && abs(dy) == 2)
	case Bishop:
		return abs(dx) == abs(dy) && g.isPathClear(from, to)
	case Queen:
		return ((dx == 0 || dy == 0) || (abs(dx) == abs(dy))) && g.isPathClear(from, to)
	case King:
		return abs(dx) <= 1 && abs(dy) <= 1
	}
	return false
}

func (g *Game) IsCheckmate(color Color) bool {
	if !g.IsInCheck(color) {
		return false
	}

	for row := range 8 {
		for col := range 8 {
			from := Position{row, col}
			piece := g.Board.At(from)
			if piece.Type != Empty && piece.Color == color {
				for toRow := range 8 {
					for toCol := range 8 {
						to := Position{toRow, toCol}
						if g.isValidMoveIgnoringCheck(from, to) {
							originalPiece := g.Board.At(to)
							g.Board.Move(from, to)
							inCheck := g.IsInCheck(color)
							g.Board.Set(from, piece)
							g.Board.Set(to, originalPiece)

							if !inCheck {
								return false
							}
						}
					}
				}
			}
		}
	}

	return true
}

func (g *Game) isValidMoveIgnoringCheck(from, to Position) bool {
	if !from.Valid() || !to.Valid() {
		return false
	}

	piece := g.Board.At(from)
	if piece.Type == Empty {
		return false
	}

	target := g.Board.At(to)
	if target.Type != Empty && target.Color == piece.Color {
		return false
	}

	return g.isValidPieceMove(piece, from, to)
}

func (g *Game) GameStatus() string {
	if g.IsCheckmate(g.CurrentTurn) {
		winner := "White"
		if g.CurrentTurn == White {
			winner = "Black"
		}
		return fmt.Sprintf("Checkmate! %s wins!", winner)
	}

	if g.IsInCheck(g.CurrentTurn) {
		return fmt.Sprintf("%s is in check!", g.CurrentTurn)
	}

	return ""
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
