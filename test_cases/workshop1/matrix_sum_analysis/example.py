import sys

input_data = sys.stdin.readlines()
# Parse input
lines = input_data.strip().split('\n')
n, m = map(int, lines[0].split())
matrix = []
for i in range(n):
    row = list(map(int, lines[i + 1].split()))
    matrix.append(row)

# Calculate column and row sums
row_sums = [sum(row) for row in matrix]
col_sums = [sum(col) for col in zip(*matrix)]

# Format output
return '\n'.join(' '.join(map(str, pair)) for pair in zip(row_sums, col_sums[:2]))
