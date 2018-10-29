# Tunnel

> Expose your local service through NAT

## What is Tunnel?

Tunnel is a simple proxy that creates a tunnel from public endpoint to a locally running service.

## Build Setup

```bash
# compile
go build tunnel.go

# show help info
./tunnel
```

## Usage

- `-server`

  - Syntax

    `./tunnel -server port1 port2`

  - Description

    Forward data between `port1` and `port2` on server.

- Client

  - Syntax

    `./tunnel -client port1 ip:port2`

  - Description

    Forward data between `127.0.0.1:port1` and `ip:port2`.
