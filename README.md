# MiniRedis – A Minimal Redis-Compatible Key-Value Store in Go

MiniRedis is a lightweight, Redis-compatible key-value store built from scratch in Go.  
It supports basic Redis commands and works directly with the official `redis-cli`.

---

## 🚀 Features

- RESP (Redis Serialization Protocol) parsing  
- Basic commands: `PING`, `SET`, `GET`, `DEL`  
- Compatible with `redis-cli`  
- Handles multiple client connections  
- Synchronous request-response communication  
- Clean modular structure for future extensions

---

## ⚙️ Running the Server

1. **Build and run**
   ```bash
   go run ./cmd/kvd

---

The server listens on port 6379 by default.

2. Connect using redis-cli
   redis-cli

---

3. **Try some commands**
```bash
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET go redis
OK
127.0.0.1:6379> GET go
"redis"




