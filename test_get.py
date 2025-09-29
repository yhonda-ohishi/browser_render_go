import requests
import json
import time

print("Testing GET endpoint...")
print("="*50)

# Simple GET request
url = "http://localhost:8091/v1/vehicle/data"
print(f"Sending GET request to {url}")
print("This will use fixed parameters (branch_id='00000000', filter_id='0')")
print("\nWaiting for response (this may take 30-60 seconds)...")

start_time = time.time()

try:
    response = requests.get(url, timeout=120)
    elapsed = time.time() - start_time

    print(f"\nResponse received in {elapsed:.1f} seconds")
    print(f"Status Code: {response.status_code}")

    if response.status_code == 200:
        data = response.json()

        # Check main response
        print(f"\nMain Response:")
        print(f"  Status: {data.get('status')}")
        print(f"  Vehicle Count: {len(data.get('data', []))}")
        print(f"  Session ID: {data.get('session_id')}")

        # Check Hono API response
        if 'hono_api' in data:
            hono = data['hono_api']
            print(f"\nHono API Response:")
            print(f"  Success: {hono.get('success')}")
            print(f"  Records Added: {hono.get('records_added')}")
            print(f"  Total Records: {hono.get('total_records')}")
            print(f"  Message: {hono.get('message')}")
        else:
            print("\nNo Hono API response in data")

        # Show sample vehicle data
        if data.get('data'):
            print(f"\nSample vehicles (first 3):")
            for i, vehicle in enumerate(data['data'][:3]):
                print(f"\n  Vehicle {i+1}:")
                print(f"    VehicleName: {vehicle.get('VehicleName')}")
                print(f"    VehicleCD: {vehicle.get('VehicleCD')}")
                print(f"    DataDateTime: {vehicle.get('DataDateTime')}")

    else:
        print(f"Error response: {response.text[:500]}")

except requests.exceptions.Timeout:
    print(f"\nRequest timed out after {time.time() - start_time:.1f} seconds")
except Exception as e:
    print(f"\nError: {e}")