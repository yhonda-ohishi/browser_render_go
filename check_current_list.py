import requests
import json

# Try the currentListAll endpoint with proper encoding
response = requests.get("https://hono-api.mtamaramu.com/api/dtakologs/currentListAll")
response.encoding = 'utf-8'

print(f"Status Code: {response.status_code}")

if response.status_code == 200:
    try:
        data = response.json()
        print(f"Total records: {len(data) if isinstance(data, list) else 'Not a list'}")

        if isinstance(data, list) and len(data) > 0:
            print("\nFirst 3 records:")
            for i, record in enumerate(data[:3]):
                print(f"\nRecord {i+1}:")
                print(f"  VehicleName: {record.get('VehicleName', 'N/A')}")
                print(f"  DataDateTime: {record.get('DataDateTime', 'N/A')}")
                print(f"  BranchName: {record.get('BranchName', 'N/A')}")
                print(f"  DriverName: {record.get('DriverName', 'N/A')}")

            # Check the latest data
            print(f"\nLast record:")
            last = data[-1]
            print(f"  VehicleName: {last.get('VehicleName', 'N/A')}")
            print(f"  DataDateTime: {last.get('DataDateTime', 'N/A')}")
            print(f"  BranchName: {last.get('BranchName', 'N/A')}")
        else:
            print("Response data:")
            print(json.dumps(data, indent=2)[:1000])
    except Exception as e:
        print(f"Error parsing response: {e}")
        print(f"Response text: {response.text[:500]}")
else:
    print(f"Error: {response.text[:500]}")