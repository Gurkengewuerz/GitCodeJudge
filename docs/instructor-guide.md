# Instructor Guide

## Initial Setup

### 1. Create Organization
1. Log into Git Server
2. Create new organization
3. Configure organization settings
    - Enable repository creation restrictions
    - Set appropriate visibility settings

### 2. Configure Webhook
1. Go to organization settings
2. Add webhook:
    - URL: `http://judge:3000/webhook`
    - Secret: Match `GITEA_WEBHOOK_SECRET`
    - Events: Select `Push`

### 3. Create Student Repositories
Use the provided script:
```bash
./scripts/create_repos.sh organization_name student_list.txt
```

Student list format:
```text
student1 email1@example.com
student2 email2@example.com
```
