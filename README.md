# Automated Programming Workshop Judge

A secure, scalable system for automatically testing student programming assignments in university workshops. This system
integrates with Gitea for submission handling and uses Docker for secure code execution.

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
- Gitea instance

## Quick Start

1. Clone the repository:

```bash
git clone git@github.com:Gurkengewuerz/GitCodeJudge.git
cd GitCodeJudge
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

## Configuration

### Environment Variables

| Variable                   | Description                             | Default                                           | Required |
|----------------------------|-----------------------------------------|---------------------------------------------------|----------|
| **Server Configuration**   |
| `SERVER_ADDRESS`           | Judge server address                    | `:3000`                                           | No       |
| `MAX_PARALLEL_JUDGES`      | Maximum parallel executions             | `5`                                               | No       |
| `TESTS_PATH`               | Path to test cases directory            | `test_cases`                                      | No       |
| **Database Configuration** |
| `DB_PATH`                  | Path to the database directory          | `database/`                                       | No       |
| `DB_TTL`                   | Database TTL in Hours. 0 means disabled | `0`                                               | No       |
| **PDF Configuration**      |
| `PDF_FOOTER_COPYRIGHT`     | Copyright text in PDF footer            | ` `                                               | No       |
| `PDF_FOOTER_GENERATEDWITH` | Generated with text in PDF footer       | `Generated with GitCodeJudge`                     | No       |
| **Gitea Configuration**    |
| `GITEA_URL`                | Gitea server URL                        | -                                                 | Yes      |
| `GITEA_TOKEN`              | Gitea API token                         | -                                                 | Yes      |
| `GITEA_WEBHOOK_SECRET`     | Webhook secret                          | -                                                 | Yes      |
| **Docker Configuration**   |
| `DOCKER_IMAGE`             | Base image for code execution           | `ghcr.io/gurkengewuerz/gitcodejudge-judge:latest` | No       |
| `DOCKER_NETWORK`           | Docker network mode                     | `none`                                            | No       |
| `DOCKER_TIMEOUT`           | Execution timeout (seconds)             | `30`                                              | No       |

### Test Case Configuration

Each programming task is defined by a `config.yaml` file in its respective directory. The configuration supports visible
and hidden test cases, task metadata, and time constraints. Example test cases can be found in [
`test_cases/`](test_cases/)

#### Directory Structure

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

#### Configuration Fields

| Field          | Type     | Description                                | Required            |
|----------------|----------|--------------------------------------------|---------------------|
| `disabled`     | boolean  | Whether the task is disabled               | No (default: false) |
| `name`         | string   | Task name                                  | Yes                 |
| `description`  | string   | Detailed task description and requirements | Yes                 |
| `start_date`   | datetime | When the task becomes available            | Yes                 |
| `end_date`     | datetime | When the task expires                      | Yes                 |
| `cases`        | array    | List of visible test cases                 | Yes                 |
| `hidden_cases` | array    | List of hidden test cases                  | No                  |

#### Test Cases

Each test case (both visible and hidden) must include:

- `input`: The input that will be provided to the student's program
- `expected`: The expected output that the program should produce

#### Important Notes:

1. Hidden test cases work exactly like visible ones but results aren't shown to students
2. Whitespace are trimmed in the expected output
3. Make sure to maintain proper indentation in the YAML file
4. Use a . for in the first row for proper YAML indentation (see [
   `test_cases/workshop1/pascal_triangle`](test_cases/workshop1/pascal_triangle/config.yaml))
5. Time constraints (`start_date` and `end_date`) use ISO 8601 format

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

5. Create a Gitea access-token for an admin user with repository _read/write_ permissions. Set the
   token as `GITEA_TOKEN`.

## Student Usage

1. Clone your assigned repository
2. Write your solution inside the `<workshop>`/`<task>`/ directory
3. Commit and push your code
4. Check the commit comments or issues for test results

Example test result:

```markdown
Test results for commit: ba54e6212c3057dc24ce4cce0682bae3b96a78b3

## ✅ All Tests Passed

### Test Results

| Test # | Task                      | Status | Time  | Details |
|--------|---------------------------|--------|-------|---------|
| 1      | workshop1/pascal_triangle | ✅     | 0.14s |         |
| 2      | workshop1/pascal_triangle | ✅     | 0.16s |         |
| 3      | workshop1/pascal_triangle | ✅     | 0.12s |         |
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
- [Maroto](https://maroto.io/#/) for PDF rendering

## Support

For support, please open an issue in the repository.
