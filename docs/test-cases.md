# Test Case Configuration

Each programming task is defined by a `config.yaml` file in its respective directory. The configuration supports visible
and hidden test cases, task metadata, and time constraints. Example test cases can be found in [
`test_cases/`](../test_cases/)

## Directory Structure

Test cases are organized by workshop and task:

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

## Configuration File Format

Each task requires a `config.yaml` file with the following structure:

```yaml
disabled: false                     # Optional: disable task
name: "Task Name"                   # Required: task name
description: |                      # Required: task description
  Detailed description of the task.
  Can be multiple lines with Markdown support.
  
start_date: "2024-01-01T00:00:00Z"  # Required: ISO 8601 format
end_date: "2024-12-31T23:59:59Z"    # Required: ISO 8601 format

cases:                              # Required: visible test cases
  - input: |
      5
      3 4
    expected: |
      7
      
hidden_cases:                       # Optional: hidden test cases
  - input: |
      10
      20
    expected: |
      30
```

### Important Notes:

1. Hidden test cases work exactly like visible ones but results aren't shown to students
2. Whitespace are trimmed in the expected output
3. Make sure to maintain proper indentation in the YAML file
4. Use a . for in the first row for proper YAML indentation (see [
   `test_cases/workshop1/pascal_triangle`](../test_cases/workshop1/pascal_triangle/config.yaml))
5. Time constraints (`start_date` and `end_date`) use ISO 8601 format


## Test Case Types

### Visible Test Cases
- Students can see both input and expected output
- Results are shown in detail
- Good for learning and debugging

### Hidden Test Cases
- Students only see if they passed or failed
- Prevents hardcoding solutions
- Tests edge cases and performance

## Best Practices

1. **Test Case Design**
    - Start with simple cases
    - Include edge cases
    - Test error handling
    - Consider performance limits

2. **Description Writing**
    - Use clear language
    - Include input/output format
    - Provide example cases
    - List constraints

3. **Time Management**
    - Set reasonable start/end dates
    - Consider timezone differences
    - Allow buffer for technical issues

## Example Test Cases

### Simple Math Task
```yaml
name: "Addition"
description: |
  Write a program that adds two numbers.
  
  Input:
  - Two space-separated integers
  
  Output:
  - Sum of the two numbers
  
cases:
  - input: "1 2"
    expected: "3"
  - input: "0 0"
    expected: "0"
    
hidden_cases:
  - input: "-1 1"
    expected: "0"
  - input: "999999 1"
    expected: "1000000"
```
