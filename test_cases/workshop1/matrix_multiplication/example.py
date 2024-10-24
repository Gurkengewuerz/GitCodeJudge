import sys
import numpy as np

input_data = sys.stdin.readlines()

# Parse input
lines = input_data.strip().split('\n')
n1, m1 = map(int, lines[0].split())
pos = 1

# Read first matrix
matrix1 = np.array([list(map(int, lines[i].split())) for i in range(pos, pos+n1)])
pos += n1

# Read second matrix dimensions
n2, m2 = map(int, lines[pos].split())
pos += 1

# Read second matrix
matrix2 = np.array([list(map(int, lines[i].split())) for i in range(pos, pos+n2)])

# Perform matrix multiplication using NumPy
result = np.matmul(matrix1, matrix2)

# Format output
return '\n'.join(' '.join(map(str, row)) for row in result)
