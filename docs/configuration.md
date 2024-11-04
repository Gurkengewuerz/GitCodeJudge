# Configuration Guide

## Server Configuration

| Variable              | Description                      | Default                 | Required |
|-----------------------|----------------------------------|-------------------------|----------|
| `SERVER_ADDRESS`      | Judge server address             | `:3000`                 | No       |
| `LOG_LEVEL`           | Log level from 0-6. 4 being Info | `4`                     | No       |
| `MAX_PARALLEL_JUDGES` | Maximum parallel executions      | `5`                     | No       |
| `TESTS_PATH`          | Path to test cases directory     | `test_cases`            | No       |
| `BASE_URL`            | Base URL for the application     | `http://localhost:3000` | No       |

## Database Configuration

| Variable  | Description                             | Default     | Required |
|-----------|-----------------------------------------|-------------|----------|
| `DB_PATH` | Path to the database directory          | `database/` | No       |
| `DB_TTL`  | Database TTL in Hours. 0 means disabled | `0`         | No       |

## PDF Configuration

| Variable                   | Description                       | Default                       | Required |
|----------------------------|-----------------------------------|-------------------------------|----------|
| `PDF_FOOTER_COPYRIGHT`     | Copyright text in PDF footer      | ` `                           | No       |
| `PDF_FOOTER_GENERATEDWITH` | Generated with text in PDF footer | `Generated with GitCodeJudge` | No       |

## Gitea Configuration

| Variable               | Description      | Default | Required |
|------------------------|------------------|---------|----------|
| `GITEA_URL`            | Gitea server URL | -       | Yes      |
| `GITEA_TOKEN`          | Gitea API token  | -       | Yes      |
| `GITEA_WEBHOOK_SECRET` | Webhook secret   | -       | Yes      |

## Docker Configuration

| Variable         | Description                   | Default                                           | Required |
|------------------|-------------------------------|---------------------------------------------------|----------|
| `DOCKER_IMAGE`   | Base image for code execution | `ghcr.io/gurkengewuerz/gitcodejudge-judge:latest` | No       |
| `DOCKER_NETWORK` | Docker network mode           | `none`                                            | No       |
| `DOCKER_TIMEOUT` | Execution timeout (seconds)   | `30`                                              | No       |

## Leaderboard & Auth Configuration

| Variable              | Description                      | Default | Required |
|-----------------------|----------------------------------|---------|----------|
| `LEADERBOARD_ENABLED` | Enable leaderboard functionality | `true`  | No       |
| `OAUTH2_ISSUER`       | The OpenID issuer URL            | -       | No       |
| `OAUTH2_CLIENT_ID`    | OAuth2 client ID                 | -       | No       |
| `OAUTH2_SECRET`       | OAuth2 client secret             | -       | No       |

### Notes

- Required variables are: `GITEA_URL`, `GITEA_TOKEN`, and `GITEA_WEBHOOK_SECRET`
- All other variables have default values and are optional
- Log level ranges from 0-6, with 4 being the default Info level
