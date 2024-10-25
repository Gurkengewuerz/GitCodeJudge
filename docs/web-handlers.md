# API Routes Documentation

This document describes all available API endpoints in the GitCodeJudge system.

## Base Routes

### Health Check
```
GET /health
```
Simple health check endpoint to verify if the service is running.

### Home Page Redirect
```
GET /
```
Redirects to `/leaderboard` using the rewrite middleware.

## Webhook Integration

### Gitea Webhook
```
POST /webhook
```
Handles repository events from the Git server. Protected by webhook secret validation middleware.

**Authentication:**
- Requires valid webhook secret in request headers
- Used for processing repository events (commits, pushes)

## Documentation

### PDF Generation
```
GET /pdf
```
Generates and serves PDF documentation for programming problems/tasks.
- Includes problem descriptions
- Test case examples
- Task requirements

## Results & Statistics

### Commit Results
```
GET /results/:commit
```
Retrieves test results for a specific commit.
- Parameter: `commit` - The commit hash to get results for
- Shows test status, execution time, and error details if any

### User Progress
```
GET /user/:username
```
Shows progress and statistics for a specific user.
- Parameter: `username` - The Gitea username
- Displays completed tasks, success rates

### Workshop Statistics
```
GET /workshop/:workshop/:task
```
Provides statistics for a specific workshop task.
- Parameters:
    - `workshop` - Workshop identifier
    - `task` - Task identifier
- Shows completion rates

### Leaderboard
```
GET /leaderboard
```
Displays the overall leaderboard.
- Task completion statistics
