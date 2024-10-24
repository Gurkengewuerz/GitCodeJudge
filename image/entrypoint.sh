#!/usr/bin/env ash
set -e

# Check if required environment variables are set
if [ -z "$JUDGE_WORKSHOP" ] || [ -z "$JUDGE_TASK" ]; then
    >&2 echo "Error: JUDGE_WORKSHOP and JUDGE_TASK environment variables must be set"
    exit 1
fi

# Define the solutions directory
SOLUTIONS_DIR="/repo/${JUDGE_WORKSHOP}/${JUDGE_TASK}"
INPUT_FILE="/judge/input.txt"

# Check if solutions directory exists
if [ ! -d "$SOLUTIONS_DIR" ]; then
    >&2 echo "Error: Solutions directory not found: $SOLUTIONS_DIR"
    exit 1
fi

# Check if input file exists
if [ ! -f "$INPUT_FILE" ]; then
    >&2 echo "Error: Input file not found: $INPUT_FILE"
    exit 1
fi

# Function to run Python files
run_python() {
    local file="$1"
    python3 "$file" < "$INPUT_FILE"
}

# Function to run Go files
run_golang() {
    local file="$1"
    # First compile the Go file
    filename=$(basename "$file")
    dirname=$(dirname "$file")
    executable="${dirname}/${filename%.*}"
    go build -o "$executable" "$file"
    # Run the compiled executable
    "$executable" < "$INPUT_FILE"
}

# Find and run all solution files
found_solutions=0

# Process Python files
for file in $(find "$SOLUTIONS_DIR" -type f -name "solution.py"); do
    run_python "$file"
    found_solutions=$((found_solutions + 1))
done

# Process Go files
for file in $(find "$SOLUTIONS_DIR" -type f -name "solution.go"); do
    run_golang "$file"
    found_solutions=$((found_solutions + 1))
done

# Check if any solutions were found and executed
if [ $found_solutions -eq 0 ]; then
    >&2 echo "Error: No solution files found in $SOLUTIONS_DIR"
    exit 1
fi
