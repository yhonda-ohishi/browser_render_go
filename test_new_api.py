import requests
import json
import time

print("Testing new browser_render with Hono API integration...")
print("="*50)

# Send request to the new server on port 8090
url = "http://localhost:8090/v1/vehicle/data"
payload = {
    "branch_id": "",  # Will be ignored, uses "00000000" internally
    "filter_id": "",  # Will be ignored, uses "0" internally
    "force_login": False
}

print(f"Sending request to {url}")
print(f"Payload: {json.dumps(payload, indent=2)}")
print("\nThis will:")
print("1. Use fixed parameters (branch_id='00000000', filter_id='0')")
print("2. Get vehicle data from the website")
print("3. Automatically send to Hono API")
print("4. Return both vehicle data and Hono API result")
print("\nWaiting for response (this may take 30-60 seconds)...")

start_time = time.time()

try:
    response = requests.post(url, json=payload, timeout=120)
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