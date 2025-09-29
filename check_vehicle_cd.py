import requests
import json

# Get fresh vehicle data from server
print("Checking VehicleCD field in server data...")
response = requests.post(
    "http://133.18.115.234:8080/v1/vehicle/data",
    headers={"Content-Type": "application/json"},
    json={"branch_id": "", "filter_id": "0", "force_login": False}
)

if response.status_code == 200:
    data = response.json()
    vehicles = data.get('data', [])

    print(f"Total vehicles: {len(vehicles)}")
    print("\nFirst 10 vehicles with VehicleCD:")

    for i, vehicle in enumerate(vehicles[:10]):
        vcd = vehicle.get('VehicleCD')
        vname = vehicle.get('VehicleName')
        dt = vehicle.get('DataDateTime')
        print(f"{i+1}. VehicleName: {vname}")
        print(f"   VehicleCD: {vcd} (type: {type(vcd).__name__})")
        print(f"   DataDateTime: {dt}")
        print()

    # Check for unique VehicleCD values
    vehicle_cds = set()
    for v in vehicles:
        vcd = v.get('VehicleCD')
        if vcd is not None:
            vehicle_cds.add(str(vcd))

    print(f"Unique VehicleCD values found: {len(vehicle_cds)}")
    if len(vehicle_cds) <= 10:
        print(f"Values: {sorted(vehicle_cds)}")
else:
    print(f"Error: {response.status_code}")