import sys
import pandas as pd

def analyze_sales_data(input_data):
    # Parse input
    lines = [line.strip() for line in input_data.strip().split('\n') if line.strip()]
    n_records = int(lines[0])

    # Create DataFrame from input data
    data = []
    for i in range(n_records):
        order_id, category, cust_type, region, amount = lines[i + 1].split(',')
        data.append({
            'order_id': order_id,
            'product_category': category,
            'customer_type': cust_type,
            'region': region,
            'amount': float(amount)
        })

    df = pd.DataFrame(data)

    # Calculate metrics
    # 1. Total revenue per product category
    category_revenue = df.groupby('product_category')['amount'].sum().sort_index()
    category_str = ' '.join(f"{cat}:{amount:.2f}" for cat, amount in category_revenue.items())

    # 2. Average order value per customer type
    customer_avg = df.groupby('customer_type')['amount'].mean().sort_index()
    customer_str = ' '.join(f"{ctype}:{avg:.2f}" for ctype, avg in customer_avg.items())

    # 3. Top performing region
    top_region = df.groupby('region')['amount'].sum().idxmax()

    # Combine results
    return f"{category_str}\n{customer_str}\n{top_region}"

# Read input
input_data = sys.stdin.read()

# Process and print output
print(analyze_sales_data(input_data))
