import requests
import json
from datetime import datetime

# First, get fresh data from the vehicle server
print("Fetching fresh vehicle data from server...")
response1 = requests.post(
    "http://133.18.115.234:8080/v1/vehicle/data",
    headers={"Content-Type": "application/json"},
    json={"branch_id": "", "filter_id": "0", "force_login": False}
)

if response1.status_code == 200:
    fresh_data = response1.json()
    vehicles = fresh_data.get('data', [])
    print(f"Current vehicles from server: {len(vehicles)}")

    if vehicles:
        # Check the dates in the fresh data
        print("\nSample of fresh data from server:")
        for i, vehicle in enumerate(vehicles[:5]):
            print(f"Vehicle {i+1}:")
            print(f"  VehicleName: {vehicle.get('VehicleName')}")
            print(f"  DataDateTime: {vehicle.get('DataDateTime')}")
            print(f"  BranchName: {vehicle.get('BranchName')}")

        # Check the date format
        if vehicles:
            sample_date = vehicles[0].get('DataDateTime')
            print(f"\nSample DataDateTime format from server: {sample_date}")

            # Count dates
            date_counts = {}
            for v in vehicles:
                dt = v.get('DataDateTime', 'None')
                if dt and dt != '<nil>':
                    # Extract just the date part
                    date_part = dt.split(' ')[0] if ' ' in dt else dt
                    date_counts[date_part] = date_counts.get(date_part, 0) + 1

            print(f"\nDate distribution in fresh data:")
            for date, count in sorted(date_counts.items(), reverse=True)[:10]:
                print(f"  {date}: {count} vehicles")
else:
    print(f"Error fetching fresh data: {response1.status_code}")

print("\n" + "="*50)
print("Now checking data in Hono API...")

# Check current data in Hono API
response2 = requests.get("https://hono-api.mtamaramu.com/api/dtakologs/currentListAll")
response2.encoding = 'utf-8'

if response2.status_code == 200:
    hono_data = response2.json()
    print(f"Total records in Hono API: {len(hono_data)}")

    # Check for today's data (2025-09-29)
    today_count = 0
    sep29_count = 0
    sep26_count = 0

    for record in hono_data:
        dt = record.get('DataDateTime', '')
        if '2025-09-29' in dt or '25/09/29' in dt:
            sep29_count += 1
        elif '2025-09-26' in dt or '25/09/26' in dt:
            sep26_count += 1

    print(f"\nRecords with 2025-09-29 (today): {sep29_count}")
    print(f"Records with 2025-09-26: {sep26_count}")

    # Show all unique dates in Hono API
    hono_dates = {}
    for record in hono_data:
        dt = record.get('DataDateTime', 'None')
        if dt:
            # Extract just the date part
            date_part = dt.split('T')[0] if 'T' in dt else dt
            hono_dates[date_part] = hono_dates.get(date_part, 0) + 1

    print(f"\nAll dates in Hono API (top 10):")
    for date, count in sorted(hono_dates.items(), reverse=True)[:10]:
        print(f"  {date}: {count} records")