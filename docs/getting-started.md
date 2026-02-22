# Getting Started

## Prerequisites

- Go 1.21 or later
- Access to both Pi-hole v6 servers:
  - `pihole.ucg.sparky.best`
  - `pihole2.ucg.sparky.best`

## Pi-hole Passwords

Each Pi-hole server has its own admin password. You can find or reset these in each Pi-hole's web interface under **Settings > Web Interface / API**.

- **pihole.ucg.sparky.best** — the primary Pi-hole server
- **pihole2.ucg.sparky.best** — the secondary Pi-hole server

## Environment Variables

Set the following environment variables with the corresponding Pi-hole admin passwords:

```sh
export HOUSE_PIHOLE1_PASSWORD="<password for pihole.ucg.sparky.best>"
export HOUSE_PIHOLE2_PASSWORD="<password for pihole2.ucg.sparky.best>"
```

You can add these to a `.env` file in the project root (it's git-ignored) and source it:

```sh
source .env
```

## Build

```sh
make build
```

This produces the `house` binary at `bin/house`.

## Usage

Temporarily disable ad blocking on both Pi-hole servers for 30 minutes:

```sh
house adblock disable
```

On success you'll see:

```
OK   pihole1: ad blocking disabled for 30 minutes
OK   pihole2: ad blocking disabled for 30 minutes
```

If a server fails, you'll see which one and why:

```
OK   pihole1: ad blocking disabled for 30 minutes
FAIL pihole2: authentication: auth request failed: connection refused
Error: one or more servers failed
```
