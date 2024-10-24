# Automated Programming Workshop Judge

A secure, scalable system for automatically testing student programming assignments in university workshops. This system integrates with Gitea for submission handling and uses Docker for secure code execution.

## Features

- **Secure Execution**: All student code runs in isolated Docker containers
- **Parallel Processing**: Handles multiple submissions simultaneously (configurable)
- **Real-time Feedback**: Students receive immediate test results on their commits
- **Scalable**: Handles large classes (100+ students) efficiently
- **Privacy**: Students can't access other students' solutions
- **Multiple Workshop Support**: Organize test cases by workshop and task
- **Flexible Test Cases**: Support for both YAML config and separate input/output files

## System Architecture

```
                   ┌─────────────┐
                   │   Gitea     │
                   │  (Git Host) │
                   └─────┬───────┘
                         │ webhook
                         ▼
┌──────────────┐   ┌─────────────┐   ┌─────────────┐
│  Test Cases  │   │   Judge     │   │   Docker    │
│  Repository  │──▶│   Server    │──▶│  Containers │
└──────────────┘   └─────────────┘   └─────────────┘
                         │
                         │ results
                         ▼
                   ┌─────────────┐
                   │  Commit     │
                   │  Comments   │
                   └─────────────┘
```

## Prerequisites

- Docker Engine 20.10+
- Docker Compose v2.0+
- Git
- Gitea instance (can be deployed with provided docker-compose)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/yourusername/workshop-judge.git
cd workshop-judge
```

2. Create `.env` file:
```env
GITEA_URL=http://gitea:3000
GITEA_TOKEN=your_gitea_token
GITEA_WEBHOOK_SECRET=your_webhook_secret
MAX_PARALLEL_JUDGES=5
DOCKER_TIMEOUT=30
```

3. Start the services:
```bash
docker-compose up -d
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_ADDRESS` | Judge server address | `:3000` |
| `MAX_PARALLEL_JUDGES` | Maximum parallel executions | `5` |
| `GITEA_URL` | Gitea server URL | Required |
| `GITEA_TOKEN` | Gitea API token | Required |
| `GITEA_WEBHOOK_SECRET` | Webhook secret | Required |
| `DOCKER_NETWORK` | Docker network name | `judge-network` |
| `DOCKER_TIMEOUT` | Execution timeout (seconds) | `30` |

### Test Case Structure

Test cases can be defined in two ways:

1. Using `config.yaml`:
```yaml
cases:
  - input: |
      3 4
      1 2 3 4
    expected: |
      10
  - input: |
      2 2
      5 6
    expected: |
      11
```

2. Using separate files:
```
workshop1/
└── matrix_multiplication/
    ├── input1.txt
    ├── output1.txt
    ├── input2.txt
    └── output2.txt
```

## Setup for Instructors

1. Create a Gitea organization for your class
2. Add webhook to the organization:
   - URL: `http://judge:3000/webhook`
   - Secret: Match `GITEA_WEBHOOK_SECRET`
   - Events: `Push`

3. Create student repositories:
```bash
./scripts/create_repos.sh organization_name student_list.txt
```

4. Add test cases to the `test_cases` directory following the structure:
```
test_cases/
├── workshop1/
│   ├── task1/
│   │   └── config.yaml
│   └── task2/
│       └── config.yaml
└── workshop2/
    └── task1/
        └── config.yaml
```

## Student Usage

1. Clone your assigned repository
2. Write your solution
3. Commit and push your code
4. Check the commit comments for test results

Example test result:
```markdown
## ✅ All Tests Passed

### Test Results
| Test # | Status | Time | Details |
|--------|--------|------|----------|
| 1 | ✅ | 0.15s | |
| 2 | ✅ | 0.12s | |
```

## Development

### Building from source:
```bash
go build -o judge ./cmd/main.go
```

### Running tests:
```bash
go test ./...
```

### Code formatting:
```bash
go fmt ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the AGPL - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gitea](https://gitea.io/) for Git hosting
- [Docker](https://www.docker.com/) for containerization
- [Go Fiber](https://gofiber.io/) for HTTP routing

## Support

For support, please open an issue in the repository.
