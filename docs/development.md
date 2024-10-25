# Development Guide

## Setup Development Environment

### Prerequisites

- Go 1.21+
- Docker
- Git

### Local Setup

1. Clone repository:
   ```bash
   git clone git@github.com:Gurkengewuerz/GitCodeJudge.git
   cd GitCodeJudge
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Building

### Build Binary

```bash
go build -o judge ./cmd/main.go
```

### Build Docker Image

```bash
docker build -t gitcodejudge -f docker/server/Dockerfile .
```

```bash
docker build -t gitcodejudge -f docker/judge/Dockerfile docker/judge
```

## Testing

### Run Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

### Format Code

```bash
go fmt ./...
```

### Run Linter

```bash
golangci-lint run
```

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/               # Private application code
│   ├── config/             # Configuration handling
│   ├── judge/              # Core judge logic
│   ├── models/             # Data models
│   └── server/             # HTTP server
└── test_cases/             # Example test cases
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
