# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CheSSH is a multiplayer chess game playable via SSH. The project consists of:
- A terminal-based chess game using Bubble Tea for the TUI
- An SSH server that hosts multiplayer chess matches 
- Automatic matchmaking system that pairs players
- Complete chess implementation with move validation, check/checkmate detection, and special moves (castling, en passant)

## Architecture

### Core Components

- **main.go**: Main entry point with Bubble Tea model for the chess TUI. Contains game state management, input handling, and UI rendering. Can run locally or as SSH server via `-port` flag.
- **server.go**: SSH server implementation using Charm's wish framework. Handles SSH sessions and integrates with the multiplayer system.
- **multiplayer.go**: Game session management, player matchmaking, and real-time updates between players. Implements the GameManager singleton for coordinating games.
- **chess.go**: Core chess engine with complete game logic including piece movement rules, board state, check/checkmate detection, castling, and en passant.

### Key Data Structures

- **Game**: Core game state with board, current turn, move history, and special move tracking
- **GameSession**: Manages a game between two players with real-time updates
- **GameManager**: Singleton that handles player queuing and matchmaking
- **Player**: Represents connected SSH session with color, game state, and update channel

## Development Commands

### Building
```bash
go build -o chessh
```

### Running Locally (Single Player)
```bash
./chessh
# or
go run .
```

### Running SSH Server
```bash
./chessh -port 2222
# or  
go run . -port 2222
```

### Testing Connection
```bash
ssh localhost -p 2222
```

## Code Organization

- Chess engine logic is completely separated from UI concerns
- Multiplayer state management uses channels for real-time updates between players
- SSH server integration uses Bubble Tea middleware from wish framework
- Game state is synchronized between players via GameUpdate messages
- Matchmaking automatically pairs players when they connect

## Key Features

- Full chess rules implementation including special moves
- Real-time multiplayer via SSH with cursor and selection synchronization
- Automatic matchmaking and game session management  
- Graceful handling of player disconnections
- Unicode chess piece symbols with color differentiation
- Game status display (check, checkmate, turn indication)
