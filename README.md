# kvx

A Redis-like key-value store written from scratch in Go. I added the `x` to the name just to add some suspense, only the `kv` part makes sense. This is not meant for production use - it's a hands-on project where I'm building my own implementation of Redis internals to understand how things work under the hood. That said, I do plan to use it in my own personal projects as I keep adding features.

## What it does

kvx is a TCP server that speaks the [RESP (Redis Serialization Protocol)](https://redis.io/docs/latest/develop/reference/protocol-spec/), which means you can connect to it using any standard Redis client or `redis-cli`. It stores data in memory and supports key expiration (TTL).

Under the hood, the server runs a single-threaded **event loop** built on Linux's epoll. Instead of spawning a goroutine per connection, it registers all client sockets with an epoll instance and loops over `EpollWait` - when a file descriptor becomes readable (new connection or incoming data), it handles it inline. This is the same model Redis itself uses. Right now the max concurrent clients is hardcoded at 20,000 - a number I picked as a starting point. I'll make it configurable once I've done some proper benchmarking. There's also a simpler synchronous server implementation available as a fallback.

## Supported commands

| Command | Example                                 | Description                                                       |
| ------- | --------------------------------------- | ----------------------------------------------------------------- |
| `PING`  | `PING` / `PING hello`                   | Returns `PONG` or echoes back your message                        |
| `SET`   | `SET key value` / `SET key value EX 60` | Store a key-value pair, optionally with expiration in seconds     |
| `GET`   | `GET key`                               | Retrieve a value by key (returns nil if expired or missing)       |
| `TTL`   | `TTL key`                               | Check remaining time to live (-1 = no expiry, -2 = doesn't exist) |

More commands coming as I need them.

## Running it

**Locally:**

```bash
make run
```

This builds and runs kvx on port `7379`. Cleanup happens automatically on exit.

**With Docker:**

```bash
make deploy
```

**Connect with redis-cli:**

```bash
redis-cli -p 7379
```

## Benchmarking

The Makefile includes targets to benchmark kvx against real Redis using `redis-benchmark`:

```bash
# kvx benchmarks
make benchmark-single-connection
make benchmark-500-concurrent

# Redis benchmarks (for comparison)
make benchmark-redis-single-connection
make benchmark-redis-500-concurrent
```

## Project structure

```
kvx/
  config/        - Server configuration (host, port)
  core/
    resp.go      - RESP protocol encoder/decoder
    store.go     - In-memory key-value store with TTL
    eval.go      - Command evaluation and execution
    cmd.go       - Command definitions
    errors.go    - Redis-style error messages
    fd.go        - File descriptor wrapper for async I/O
  server/
    async_tcp.go - Epoll-based async TCP server
    sync_tcp.go  - Synchronous TCP server (alternative)
    io.go        - Command reading and response writing
```

## Some implementation notes

- The RESP protocol implementation handles all the standard types: simple strings, errors, integers, bulk strings, and arrays (including nested ones).
- Key expiration is checked lazily on access (GET/TTL), similar to how Redis does it. No background goroutine sweeping expired keys yet.
- The async server uses raw syscalls and epoll to avoid the overhead of Go's net package. The file descriptor wrapper implements `io.ReadWriter` so it still plays nicely with Go interfaces.
- No persistence - everything lives in memory and is gone when the server stops. I might add AOF or snapshotting later.
