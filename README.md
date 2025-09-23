# CheSSH

A multiplayer chess game playable via SSH.

> [!IMPORTANT]
> It's pronounced like _"chess-ess-aych"_, or _"chesh"_ in your best Sean Connery impression.

## Play the game

```
ssh chessh.app.imjasonh.com
```

![Screenshot of CheSSH gameplay](screenshot.png)

You may need to wait for another player to play against, or join from another terminal if you want to play with yourself.

Players are automatically matched when they connect to the SSH server.

When an opponent disconnects, you win!

## Running locally

```bash
go run ./ --local
```

then

```bash
ssh localhost -p 2222
```

## Deploying

```bash
cd iac/
terraform init
terraform apply
```

This will create:
- A container image containing the app's code
- GCE VM instances running the app container
- TCP Load Balancer for direct SSH access on port 22
- Health checks and auto-healing for high availability
- DNS records pointing to the load balancer
- SSH host key management via Secret Manager
