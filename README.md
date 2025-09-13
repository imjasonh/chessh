# CheSSH

A multiplayer chess game playable via SSH.

## Usage

```bash
# Run SSH server
go run ./ -port 2222

# Connect to server
ssh localhost -p 2222

# (In another terminal) Connect another player session
ssh localhost -p 2222
```

Players are automatically matched when they connect to the SSH server.

When an opponent disconnects, you win!
