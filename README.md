# WiseGuard

WiseGuard is a TCP server protected from DDoS attacks using the Proof of Work (PoW) mechanism. After successfully solving the PoW challenge, the server sends a random quote to the client.

## Features

- TCP server with DDoS protection
- Proof of Work protocol implementation (based on SHA-256)
- Dynamic PoW difficulty adjustment based on server load
- Graceful shutdown
- Configurable timeouts and connection limits
- Efficient error handling and logging
- Docker containerization
- Built-in load testing tool

## Proof of Work and DDoS Protection

### Algorithm
The PoW algorithm uses HashCash based on SHA-256. Clients must find a nonce where the hash of the concatenation prefix + nonce begins with a specified number of zero bits.

### Advantages of the Chosen Approach
- Simple implementation
- Efficient server-side verification
- Scalable difficulty
- Cryptographic strength of SHA-256

### Protection Mechanism
1. Dynamic Difficulty: Server automatically adjusts PoW difficulty based on current load
2. Timeouts: All operations have configurable timeouts
3. Connection Limiting: Configurable limit on simultaneous connections
4. Solution Validation: Strict verification of all PoW solutions

## Requirements

- Go 1.22 or higher
- Docker and Docker Compose (for containerized deployment)
- Make (for using Makefile)

## Installation and Running

### Local Setup

1. Clone the repository:
```bash
git clone git@github.com:aturan23/WiseGuard.git
cd WiseGuard
```

2. Build the project:
```bash
make build
```

3. Run the server:
```bash
make run-server
```

4. In another terminal, run the client:
```bash
make run-client
```

### Docker Setup

1. Build Docker images:
```bash
make docker-build
```

2. Run containers:
```bash
make docker-run
```

## Load Testing

WiseGuard includes a tool for load testing and DDoS attack simulation.

### Running Tests

```bash
# Standard test (100 clients, 30 seconds)
make run-ddos

# Heavy test (1000 clients, 60 seconds)
make run-ddos-heavy

# Custom test
./bin/ddos -clients 500 -duration 45s -server localhost:8080
```

### Performance Under Different Loads

#### Normal Load (100 concurrent clients)
```
Total Requests: ~1000
Successful Requests: ~800
Success Rate: ~85%
Average Latency: ~2.6s
Average RPS: ~32
```

#### Heavy Load (1000 concurrent clients)
```
Total Requests: ~3750
Successful Requests: ~800
Success Rate: ~22%
Average Latency: ~3.8s
Average RPS: ~63
```

### Understanding the Results

The system demonstrates effective DDoS protection:

1. **Under Normal Load:**
   - High success rate (~85%)
   - Lower latency (~2.6s)
   - Stable RPS
   - Most legitimate requests are processed

2. **Under Heavy Load (DDoS Simulation):**
   - Lower success rate (~22%) - intentionally rejecting excess traffic
   - Slightly increased latency (~3.8s)
   - Maintained stable service for legitimate requests
   - Successfully prevented system overload

This behavior shows that the PoW mechanism effectively:
- Throttles excessive traffic
- Maintains service for legitimate users
- Scales resource usage under load
- Prevents server overload

### Protection Mechanisms in Action

During heavy load:
1. **Dynamic PoW Difficulty**
   - Automatically increases with load
   - Makes attack more computationally expensive

2. **Connection Management**
   - Limits concurrent connections
   - Enforces timeouts
   - Prevents resource exhaustion

3. **Resource Protection**
   - Memory usage remains stable
   - CPU load is distributed to clients
   - Network bandwidth is preserved

### Interpreting Results

Good performance indicators vary by scenario:

**Normal Operation:**
- Success Rate > 80%
- Latency < 3s
- Stable RPS

**Under Attack:**
- Success Rate: 20-30% is optimal
- Latency increase < 50%
- Consistent successful request count

## Development

### Available Commands

- `make build` - Build the project
- `make test` - Run all tests with coverage
- `make test-unit` - Run unit tests only
- `make test-integration` - Run integration tests
- `make clean` - Clean build artifacts
- `make fmt` - Format code
- `make lint` - Run linter
- `make bench` - Run benchmarks
- `make race` - Check for race conditions
- `make run-server` - Run server locally
- `make run-client` - Run client locally
- `make docker-build` - Build Docker images
- `make docker-run` - Run Docker containers
- `make docker-stop` - Stop Docker containers
- `make run-ddos` - Run DDoS simulation
- `make run-ddos-heavy` - Run heavy DDoS simulation

### Project Structure

```
.
├── cmd/
│   ├── client/        # Client entry point
│   └── server/        # Server entry point
├── pkg/
│   ├── client/        # Client logic
│   ├── config/        # Configuration
│   ├── logger/        # Logging
│   ├── pow/          # Proof of Work implementation
│   ├── protocol/      # Network protocol
│   ├── quotes/        # Quotes service
│   ├── server/        # Server logic
│   └── utils/         # Helper functions
├── integration/       # Integration tests
├── configs/          # Configuration files
├── Dockerfile.client  # Client Dockerfile
├── Dockerfile.server  # Server Dockerfile
├── docker-compose.yml # Docker Compose config
└── Makefile          # Make commands
```

### Configuration

Application configuration through environment variables or yaml files:

#### Server Configuration
```yaml
# Server settings
SERVER_ADDRESS=:8080
SERVER_READ_TIMEOUT=5s
SERVER_WRITE_TIMEOUT=5s
SERVER_SHUTDOWN_TIMEOUT=10s
SERVER_MAX_CONNECTIONS=1000

# Logging
LOG_LEVEL=debug
LOG_PRETTY=true
LOG_JSON=false

# PoW settings
POW_INITIAL_DIFFICULTY=4
POW_MAX_DIFFICULTY=8
POW_CHALLENGE_TTL=5m
POW_ADJUST_INTERVAL=10s
```

#### Client Configuration
```yaml
# Client settings
CLIENT_SERVER_ADDRESS=localhost:8080
CLIENT_CONNECT_TIMEOUT=5s
CLIENT_READ_TIMEOUT=5s
CLIENT_WRITE_TIMEOUT=5s
CLIENT_MAX_ATTEMPTS=3
CLIENT_RETRY_DELAY=1s
CLIENT_MAX_RETRY_DELAY=30s

# Logging
LOG_LEVEL=debug
LOG_PRETTY=true
LOG_JSON=false
```

## Protocol

The message exchange protocol between client and server:

1. Client connects to server
2. Server sends PoW challenge (prefix, difficulty, nonce)
3. Client solves challenge and sends solution
4. Server verifies solution:
   - If successful, sends quote
   - If failed, sends error

### Message Format

Each message has a header (8 bytes):
- Version (1 byte)
- Type (1 byte)
- Flags (2 bytes)
- Length (4 bytes)

The header is followed by a JSON payload.

### Message Types
- `TypeChallenge (1)`: Server sends PoW challenge
- `TypeSolution (2)`: Client sends solution
- `TypeQuote (3)`: Server sends quote
- `TypeError (4)`: Error message

### Payload Examples

#### Challenge
```json
{
    "prefix": "abc123",
    "difficulty": 4,
    "nonce": "xyz789",
    "expires_at": "2024-01-01T12:00:00Z"
}
```

#### Solution
```json
{
    "prefix": "abc123",
    "solution": "def456",
    "nonce": "xyz789"
}
```

#### Quote
```json
{
    "text": "The journey of a thousand miles begins with one step.",
    "author": "Lao Tzu"
}
```

## Testing

The project is covered by unit tests and integration tests. To run:

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests only
make test-integration

# Generate coverage report
make test-coverage
```

### Test Coverage
- Protocol implementation
- PoW logic
- Client/Server interaction
- Configuration handling
- Error scenarios
- Load testing

## Logging

The application uses structured logging via zerolog. Log levels are configurable:

- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `LOG_PRETTY` - Human-readable log format
- `LOG_JSON` - JSON log format

### Log Example
```json
{
    "level": "info",
    "timestamp": "2024-11-08T12:00:00Z",
    "component": "server",
    "message": "client connected",
    "remote_addr": "127.0.0.1:12345"
}
```

## Security Considerations

1. **DoS Protection**
   - PoW-based challenge-response
   - Connection limits
   - Timeouts
   - Memory usage controls

2. **Resource Management**
   - Connection pooling
   - Garbage collection
   - Buffer size limits

3. **Error Handling**
   - Graceful degradation
   - Proper error reporting
   - Secure error messages

