import requests
import json

# Try different branch_id values
branch_ids = ["", "0", "1", "2", "3", "all", "*"]

print("Testing different branch_id values...")
print("="*50)

for branch_id in branch_ids:
    print(f"\nTrying branch_id: '{branch_id}'")

    response = requests.post(
        "http://133.18.115.234:8080/v1/vehicle/data",
        headers={"Content-Type": "application/json"},
        json={"branch_id": branch_id, "filter_id": "0", "force_login": False}
    )

    print(f"  Status: {response.status_code}")

    if response.status_code == 200:
        data = response.json()
        vehicles = data.get('data', [])
        print(f"  Vehicles returned: {len(vehicles)}")

        if vehicles:
            # Check sample vehicle
            sample = vehicles[0]
            print(f"  Sample vehicle:")
            print(f"    VehicleName: {sample.get('VehicleName')}")
            print(f"    BranchCD: {sample.get('BranchCD')}")
            print(f"    BranchName: {sample.get('BranchName')}")
            print(f"    VehicleCD: {sample.get('VehicleCD')}")

            # Check for unique branch codes
            branch_codes = set()
            for v in vehicles:
                bc = v.get('BranchCD')
                if bc:
                    branch_codes.add(str(bc))

            if branch_codes:
                print(f"  Unique BranchCD values found: {sorted(branch_codes)[:10]}")
    else:
        print(f"  Error: {response.text[:100]}")

print("\n" + "="*50)
print("\nNow checking if we can get branch list...")

# Try to get branch list
branch_endpoints = [
    "/v1/branch/list",
    "/v1/branches",
    "/v1/branch/data",
    "/v1/master/branch"
]

for endpoint in branch_endpoints:
    url = f"http://133.18.115.234:8080{endpoint}"
    print(f"\nTrying: {url}")

    try:
        response = requests.get(url)
        if response.status_code == 200:
            print(f"  Success! Status: {response.status_code}")
            data = response.json()
            print(f"  Response sample: {str(data)[:200]}")
            break
        else:
            print(f"  Status: {response.status_code}")
    except:
        try:
            response = requests.post(url, json={})
            if response.status_code == 200:
                print(f"  Success with POST! Status: {response.status_code}")
                data = response.json()
                print(f"  Response sample: {str(data)[:200]}")
                break
            else:
                print(f"  POST Status: {response.status_code}")
        except:
            print(f"  Failed")