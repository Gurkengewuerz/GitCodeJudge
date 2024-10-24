import sys

def generate_pascal(n):
    # Generate Pascal's triangle rows
    triangle = [[1]]
    for i in range(1, n):
        row = [1]
        for j in range(1, i):
            row.append(triangle[i-1][j-1] + triangle[i-1][j])
        row.append(1)
        triangle.append(row)
    return triangle

def format_triangle(triangle):
    # Convert all numbers to strings and get max width
    str_triangle = [[str(num) for num in row] for row in triangle]
    max_width = len(max([num for row in str_triangle for num in row], key=len))

    # Calculate the width of the last row to center all rows
    last_row_width = len(triangle[-1]) * (max_width + 1) - 1

    # Format each row with proper spacing
    result = []
    for row in str_triangle:
        # Center the numbers in their slots
        formatted_nums = [num.center(max_width) for num in row]
        # Join with space and center the entire row
        row_str = ' '.join(formatted_nums)
        result.append(row_str.center(last_row_width))

    return '\n'.join(result)

def solve(input_data):
    n = int(input_data.strip())
    triangle = generate_pascal(n)
    return format_triangle(triangle)


input_data = "\n".join(sys.stdin.readlines())
print(solve(input_data))
