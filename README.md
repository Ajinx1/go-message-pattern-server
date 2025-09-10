# go-message-pattern-server

A lightweight Go library providing **NestJS-style Message Pattern** handling over TCP for **fast, internal service-to-service communication**.

Unlike traditional REST APIs, this library keeps **long-lived TCP connections** with **JSON-based message routing**, providing **lower latency** and **cleaner microservice communication** patterns.

---

## Features

* 🚀 **NestJS-like message patterns**: Use `MessagePattern` handlers for simple routing
* 🔄 **Long-lived TCP connections**: Persistent connections with heartbeat support
* 📊 **Built-in metrics**: Requests, errors, active connections, processing time
* ⚡ **Rate limiting**: Per-connection request limits
* 🔁 **Retry logic**: Automatic retries for failed handlers
* 🛑 **Graceful shutdown**: Clean shutdown with context timeout
* 🧠 **Simple API**: Minimal boilerplate, plug-and-play handlers

---

## Installation

```bash
go get github.com/Ajinx1/go-message-pattern-server
```

---

## Quick Start

```go
package main

import (
 "context"
 "encoding/json"
 "fmt"
 "time"
 "github.com/Ajinx1/go-message-pattern-server/rpc"
)

func main() {
 ctx := context.Background()

 // Server configuration
 config := &rpc.Config{
  Addr:              ":4061",
  Timeout:           30 * time.Second,
  MaxConnections:    1000,
  RateLimitPerSec:   100,
  RateLimitBurst:    200,
  RetryAttempts:     3,
  RetryDelay:        500 * time.Millisecond,
  HeartbeatInterval: 10 * time.Second,
  HeartbeatTimeout:  30 * time.Second,
 }

 // Create server
 server := rpc.NewServer(config, nil)
 wrapper := rpc.NewServerWrapper(server)

 // Register message patterns
 wrapper.MessagePattern("ping", func(data json.RawMessage) (interface{}, error) {
  return "pong", nil
 })

 wrapper.MessagePattern("echo", func(data json.RawMessage) (interface{}, error) {
  var payload map[string]interface{}
  json.Unmarshal(data, &payload)
  return payload, nil
 })

 // Start server
 go func() {
  if err := server.Start(); err != nil {
   fmt.Printf("Server error: %v\n", err)
  }
 }()

 // Wait for interrupt
 <-ctx.Done()
 server.Shutdown(context.Background())
}
```

---

## Client Example

```go
// Send a message to the server
conn, _ := net.Dial("tcp", "localhost:4061")
defer conn.Close()

msg := `{"pattern":"ping","data":{}}`
conn.Write([]byte(msg + "\n"))
```

---

## Configuration

| Option            | Type          | Default   | Description                          |
| ----------------- | ------------- | --------- | ------------------------------------ |
| Addr              | string        | `":8080"` | TCP address to listen on             |
| Timeout           | time.Duration | `30s`     | Request handling timeout             |
| MaxConnections    | int           | `1000`    | Max concurrent connections           |
| RateLimitPerSec   | int           | `100`     | Requests per second per connection   |
| RateLimitBurst    | int           | `200`     | Burst limit for rate limiting        |
| RetryAttempts     | int           | `3`       | Retries for failed handlers          |
| RetryDelay        | time.Duration | `500ms`   | Delay between retries                |
| HeartbeatInterval | time.Duration | `10s`     | How often to send heartbeat messages |
| HeartbeatTimeout  | time.Duration | `30s`     | Timeout for heartbeat response       |

---

## Metrics Endpoint Example (Optional)

You can expose metrics via Fiber:

```go
app.Get("/metrics", func(c *fiber.Ctx) error {
 metrics := server.GetMetrics()
 return c.JSON(metrics)
})
```

---

## Why Use This Instead of REST?

* REST opens **new connections** for every request → more overhead
* This library keeps **a single TCP connection** → faster, fewer round trips
* Ideal for **internal microservices** that need low-latency communication

---

---

## License

MIT

---