# Student Guide

## Getting Started

1. Get your repository credentials from your instructor
2. Clone your repository:
   ```bash
   git clone <repository-url>
   ```
3. Set up your SSH key in Git for access

## Directory Structure

Your repository should follow this structure:

```
your-repo/
├── workshop1/
│   ├── task1/
│   │   └── solution.py  # or other file types
│   └── task2/
│       └── solution.py
└── workshop2/
    └── task1/
        └── solution.go
```

## Working on Tasks

### Method 1: Local Development

1. Write your solution locally
2. Test with provided test cases
3. Commit and push:
   ```bash
   git add .
   git commit -m "Solve task1"
   git push
   ```

### Method 2: SSHContainer Development

1. Connect to development environment:
   ```bash
   ssh -p 2222 git@<sshcontainer-host>
   ```
2. Write and test your solution
3. Commit and push from container

## Understanding Test Results

Test results appear as commit status with a link to the detailed results.:

```markdown
## ✅ All Tests Passed

### Test Results

| Test # | Task              | Status | Time  | Details |
|--------|-------------------|--------|-------|---------|
| 1      | workshop1/task1   | ✅     | 0.14s |         |
| 2      | workshop1/task1   | ✅     | 0.16s |         |
```

## Tips

- Check task deadlines in test case descriptions
- Use meaningful commit messages
- Test your code with edge cases
- Ask for help if you're stuck
