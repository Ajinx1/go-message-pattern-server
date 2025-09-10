# go-message-pattern-server

A lightweight Go library providing NestJS-style **Message Pattern** handling over TCP.  
Designed for **service-to-service communication** using JSON payloads. Supports message pattern routing, heartbeat, metrics, rate limiting, and retries.

---

## Features

- NestJS-like message pattern routing over TCP
- JSON payloads for requests and responses
- Heartbeat support for long-lived connections
- Metrics: total requests, errors, heartbeats, active connections, processing time
- Rate limiting per connection
- Retry logic for handlers
- Graceful shutdown

---

## Installation

```bash
go get github.com/Ajinx1/go-message-pattern-server
