# Configuration Guide

### Environment Variables

| Variable                   | Description                             | Default                                           | Required |
|----------------------------|-----------------------------------------|---------------------------------------------------|----------|
| **Server Configuration**   |
| `SERVER_ADDRESS`           | Judge server address                    | `:3000`                                           | No       |
| `LOG_LEVEL`                | Log level from 0-6. 4 being Info        | `4`                                               | No       |
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
