import sys

def solve_matrix_sum(input_data):
    # Parse input
    lines = input_data.strip().split('\n')
    if not lines:
        return ""

    # Parse first line for dimensions
    N, M = map(int, lines[0].split())

    # Read matrix
    matrix = []
    for i in range(1, N+1):
        nums = list(map(int, lines[i].split()))
        matrix.append(nums)

    result = []

    # Process each row
    for i in range(N):
        # Calculate row sum
        row_sum = sum(matrix[i])

        # Calculate first two columns sum up to current row (inclusive)
        col_sum = 0
        num_cols = min(2, M)  # Use at most 2 columns
        for r in range(i+1):
            for c in range(num_cols):
                col_sum += matrix[r][c]

        result.append(f"{row_sum} {col_sum}")

    return '\n'.join(result)

input_data = sys.stdin.read()
print(solve_matrix_sum(input_data))
