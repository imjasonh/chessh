# CheSSH

A multiplayer chess game playable via SSH.

> [!IMPORTANT]
> It's pronounced like _"chess-ess-aych"_, or _"chesh"_ in your best Sean Connery impression.

## Play the game

```
ssh chessh.imjasonh.dev
```

You may need to wait for another player to play against, or join from another terminal to play against yourself.

Players are automatically matched when they connect to the SSH server.

When an opponent disconnects, you win!

## Deploying

```bash
cd iac/
terraform init
terraform apply
```

This will create a GKE Autopilot cluster, Cloud Run service, host key secret, GCLB, routing rules, etc., to be able to host the server on the internet.
