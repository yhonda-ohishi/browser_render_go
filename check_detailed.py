import requests
import json
from datetime import datetime

# Get all data from currentListAll endpoint
response = requests.get("https://hono-api.mtamaramu.com/api/dtakologs/currentListAll")
response.encoding = 'utf-8'

print(f"Status Code: {response.status_code}")

if response.status_code == 200:
    data = response.json()
    print(f"Total records: {len(data) if isinstance(data, list) else 'Not a list'}")

    if isinstance(data, list) and len(data) > 0:
        # Group by DataDateTime to see distribution
        date_counts = {}
        for record in data:
            dt = record.get('DataDateTime', 'Unknown')
            if dt not in date_counts:
                date_counts[dt] = 0
            date_counts[dt] += 1

        print("\nData distribution by DateTime:")
        for dt, count in sorted(date_counts.items())[:10]:  # Show first 10 dates
            print(f"  {dt}: {count} records")

        # Check for recent data (today's date)
        today_str = datetime.now().strftime("%Y-%m-%d")
        recent_count = 0
        for record in data:
            if record.get('DataDateTime', '').startswith(today_str) or \
               record.get('DataDateTime', '').startswith("2025"):
                recent_count += 1

        print(f"\nRecords with 2025 date: {recent_count}")

        # Show unique VehicleNames
        vehicle_names = set()
        for record in data:
            vehicle_names.add(record.get('VehicleName', 'Unknown'))

        print(f"\nUnique vehicles: {len(vehicle_names)}")
        print("Sample vehicle names:")
        for name in list(vehicle_names)[:5]:
            print(f"  {name}")

        # Check the last 5 records
        print("\nLast 5 records:")
        for i, record in enumerate(data[-5:]):
            print(f"\nRecord {len(data)-4+i}:")
            print(f"  VehicleName: {record.get('VehicleName')}")
            print(f"  DataDateTime: {record.get('DataDateTime')}")
            print(f"  ID: {record.get('id')}")
            print(f"  CreatedAt: {record.get('CreatedAt')}")
else:
    print(f"Error: {response.text[:500]}")