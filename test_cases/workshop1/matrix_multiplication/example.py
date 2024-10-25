import sys
import numpy as np

def solve_matrix_multiplication(input_data):
    # Parse input - filter out empty lines
    lines = [line for line in input_data.strip().split('\n') if line.strip()]

    # Get dimensions of first matrix
    N1, M1 = map(int, lines[0].split())
    current_line = 1

    # Read first matrix
    matrix_A = []
    for i in range(N1):
        row = list(map(int, lines[current_line + i].split()))
        matrix_A.append(row)
    current_line += N1

    # Get dimensions of second matrix
    N2, M2 = map(int, lines[current_line].split())
    current_line += 1

    # Read second matrix
    matrix_B = []
    for i in range(N2):
        row = list(map(int, lines[current_line + i].split()))
        matrix_B.append(row)

    # Convert to numpy arrays
    A = np.array(matrix_A)
    B = np.array(matrix_B)

    # Perform matrix multiplication
    C = np.matmul(A, B)

    # Format output
    result = []
    for row in C:
        result.append(' '.join(map(str, row)))

    return '\n'.join(result)

# Read input
input_data = "\n".join(sys.stdin.readlines())

# Process and print output
print(solve_matrix_multiplication(input_data))
