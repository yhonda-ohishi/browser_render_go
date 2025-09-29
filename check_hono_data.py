import requests

# Get data from Hono API
response = requests.get("https://hono-api.mtamaramu.com/api/dtakologs")

if response.status_code == 200:
    data = response.json()
    print(f"Total records in Hono API: {len(data)}")

    # Check if data exists
    if len(data) > 0:
        # Show first few records
        print("\nFirst 3 records:")
        for i, record in enumerate(data[:3]):
            if isinstance(record, dict):
                print(f"\nRecord {i+1}:")
                print(f"  VehicleName: {record.get('VehicleName', 'N/A')}")
                print(f"  DataDateTime: {record.get('DataDateTime', 'N/A')}")
                print(f"  BranchName: {record.get('BranchName', 'N/A')}")

        # Show last record timestamp
        if len(data) > 0:
            last_record = data[-1]
            print(f"\nLast record:")
            print(f"  VehicleName: {last_record.get('VehicleName', 'N/A')}")
            print(f"  DataDateTime: {last_record.get('DataDateTime', 'N/A')}")
    else:
        print("No data found in Hono API")
else:
    print(f"Error getting data: {response.status_code}")
    print(response.text[:500])