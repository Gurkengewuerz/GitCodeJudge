# Quick Start Guide

## Prerequisites

- Docker Engine 20.10+
- Docker Compose v2.0+
- Git
- Gitea instance
- Traefik

## Installation

1. Clone the repository:
```bash
git clone git@github.com:Gurkengewuerz/GitCodeJudge.git
cd docker
```

2. Create `.env` file:
```env
CONTAINER_DIR=./data
VIRTUAL_HOST=judge.example.com
GITEA_URL=https://gitea.example.com
GITEA_TOKEN=your_gitea_token
GITEA_WEBHOOK_SECRET=your_webhook_secret
```

3. Start the services:
```bash
docker-compose up -d
```

## Verify Installation

1. Check service status:
```bash
docker-compose ps
```

2. View logs:
```bash
docker-compose logs -f
```

## Next Steps

- [Configure your environment](configuration.md)
- [Set up test cases](test-cases.md)
- [Read the instructor guide](instructor-guide.md)
