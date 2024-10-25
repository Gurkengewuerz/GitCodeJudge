import pandas as pd
import re
from collections import Counter
from statistics import mode

def analyze_text(input_data):
    # Parse input
    lines = [line.strip() for line in input_data.strip().split('\n') if line.strip()]
    n_lines = int(lines[0])
    text = ' '.join(lines[1:n_lines+1])
    k = int(lines[n_lines+1])

    # Clean and process text
    # Remove punctuation except hyphens, convert to lowercase
    cleaned_text = re.sub(r'[.,!?;:"()\[\]{}]', '', text.lower())

    # Split into words and filter out short words
    words = [word for word in cleaned_text.split() if len(word) >= 2]

    # Create DataFrame for word frequency analysis
    word_freq = pd.DataFrame(Counter(words).items(), columns=['word', 'frequency'])
    word_freq = word_freq.sort_values(['frequency', 'word'], ascending=[False, True])

    # Get top K words
    top_k = word_freq.head(k)
    top_k_str = '\n'.join(f"{row['word']}: {row['frequency']}" for _, row in top_k.iterrows())

    # Calculate word length statistics
    word_lengths = pd.Series([len(word) for word in words])
    length_stats = {
        'mean': round(word_lengths.mean(), 2),
        'median': round(word_lengths.median(), 2),
        'mode': round(float(mode(word_lengths)), 2)
    }
    stats_str = '\n'.join(f"{key}: {value:.2f}" for key, value in length_stats.items())

    # Find longest words
    max_length = max(len(word) for word in words)
    longest_words = sorted(set(word for word in words if len(word) == max_length))
    longest_str = ', '.join(longest_words)

    # Combine results
    return f"{top_k_str}\n---\n{stats_str}\n---\n{longest_str}"

# For testing
if __name__ == "__main__":
    import sys
    input_data = sys.stdin.read()
    print(analyze_text(input_data))
