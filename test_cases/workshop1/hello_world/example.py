import pandas as pd

def generate_greetings(input_data):
    # Parse input
    lines = [line.strip() for line in input_data.strip().split('\n') if line.strip()]
    n_people = int(lines[0])

    # Create DataFrame
    data = []
    for i in range(n_people):
        name, age = lines[i + 1].split()
        data.append({'name': name, 'age': int(age)})

    df = pd.DataFrame(data)

    # Define age category function
    def get_age_label(age):
        if age < 13:
            return " (child)"
        elif age <= 19:
            return " (teenager)"
        return ""

    # Apply formatting using pandas
    df['age_label'] = df['age'].apply(get_age_label)
    df['greeting'] = df.apply(
        lambda row: f"Hello, {row['name']}! You are {row['age']} years old.{row['age_label']}",
        axis=1
    )

    # Return formatted greetings
    return '\n'.join(df['greeting'].tolist())

# For testing
if __name__ == "__main__":
    import sys
    input_data = sys.stdin.read()
    print(generate_greetings(input_data))
