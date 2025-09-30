# telemetry-relay

![CI](https://github.com/mventrelli/telemetry-relay/actions/workflows/ci.yml/badge.svg)
![Go Version](https://img.shields.io/badge/go-1.25%2B-blue)

A small Go service that simulates an **operations-style telemetry ingestion pipeline**:

* Listens for telemetry packets over **UDP**
* Parses JSON into a normalized Go struct
* Exposes `/healthz` and `/metrics` endpoints (Prometheus-compatible)
* Optionally forwards packets to another HTTP API (e.g. a Flask service)
* Includes a sender utility to generate sample packets

This project was built to demonstrate systems engineering and Go development skills relevant to spacecraft/launch operations.

## Quickstart

### Prerequisites

* Go 1.25+ installed
* Git (for cloning)

### Clone

```bash
git clone https://github.com/mventrelli/telemetry-relay.git
cd telemetry-relay
```

### Run the Relay

Start the telemetry relay service:

```bash
go run ./cmd/relay
```

By default it:
* Listens for UDP packets on **:9000**
* Serves health + metrics on **:8080**

### Run the Sender

In another terminal, start the sample telemetry sender:

```bash
go run ./cmd/sender
```

The sender emits a JSON packet every 500ms with fields like sequence number, pump RPM, and tank temperature.

### Verify It's Working

* **Health check:** http://localhost:8080/healthz → returns `200 OK`
* **Metrics:** http://localhost:8080/metrics → Prometheus metrics. Look for `telemetry_ingested_total` increasing as packets arrive

## Example Log Output

Relay logs when packets are received:

```
2025/09/30 14:30:05 received packet seq=206 source=stage1 values=map[pump_rpm:5023.0 tank_temp_c:12.1]
2025/09/30 14:30:06 received packet seq=207 source=stage1 values=map[pump_rpm:5061.7 tank_temp_c:10.2]
```

## Configuration

The relay can be customized via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `UDP_ADDR` | `:9000` | UDP listen address |
| `HTTP_ADDR` | `:8080` | HTTP listen address (health + metrics) |
| `FORWARD_URL` | *(empty)* | Optional HTTP endpoint to forward packets |
| `WORKERS` | `4` | Number of forwarding worker goroutines |
| `QUEUE_SIZE` | `1024` | Size of job queue |

Example (custom ports + forward URL):

```bash
UDP_ADDR=":9900" HTTP_ADDR=":8088" FORWARD_URL="http://localhost:5000/api/telemetry" go run ./cmd/relay
```

## Project Layout

```
telemetry-relay/
├── cmd/
│   ├── relay/      # Main telemetry relay service
│   └── sender/     # Test telemetry generator
├── internal/
│   ├── telemetry/      # Packet struct definition
│   └── observability/  # Metrics + health router
├── .github/
│   └── workflows/
│       └── ci.yml      # GitHub Actions CI/CD
└── README.md           # This file
```

## CI/CD

* GitHub Actions workflow (`.github/workflows/ci.yml`) builds and tests the project on every push

## Why This Project?

This repo demonstrates skills relevant to a **Software Engineer II – Operations Software** role:

- **Go (Golang)** networking and concurrency (UDP ingest, worker pool)
- **Telemetry ingestion** and normalization for ops workflows
- **Observability**: `/healthz`, Prometheus **/metrics**, structured logs
- **Reliability** fundamentals: bounded queues, graceful shutdown
- **CI/CD**: GitHub Actions build + test on every push

## Author

Matthew Ventrelli — [email](mailto:mventrelli99@gmail.com)


## License

MIT © 2025 Matthew Ventrelli
