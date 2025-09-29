import requests
import json
import hashlib
import re
from datetime import datetime

def convert_to_number(value):
    """Convert string to number if possible, otherwise return original value"""
    if value is None or value == "<nil>" or value == "":
        return 0
    if isinstance(value, str):
        try:
            return float(value)
        except ValueError:
            return 0
    return value

def extract_vehicle_code(vehicle_name):
    """Extract a unique code from vehicle name or generate one"""
    if not vehicle_name:
        return 0

    # Try to extract numbers from vehicle name
    # e.g., "����800��11" -> extract "80011" or similar
    numbers = re.findall(r'\d+', vehicle_name)
    if numbers:
        # Combine all numbers to create a unique code
        combined = ''.join(numbers)
        try:
            return int(combined) % 2147483647  # Keep it within integer range
        except:
            pass

    # If no numbers found, generate a hash-based ID
    hash_val = hashlib.md5(vehicle_name.encode('utf-8', errors='ignore')).hexdigest()
    return int(hash_val[:8], 16) % 2147483647

def format_datetime(dt_string):
    """Format datetime to match what the API expects"""
    if not dt_string or dt_string == "<nil>":
        return None
    try:
        if dt_string.startswith("20"):
            dt_string = dt_string[2:]
        return dt_string
    except:
        return dt_string

# Step 1: Get fresh vehicle data
print("Fetching fresh vehicle data...")
response1 = requests.post(
    "http://133.18.115.234:8080/v1/vehicle/data",
    headers={"Content-Type": "application/json"},
    json={"branch_id": "", "filter_id": "0", "force_login": False}
)

if response1.status_code != 200:
    print(f"Error fetching data: {response1.status_code}")
    exit(1)

data = response1.json()
vehicles = data.get('data', [])
print(f"Retrieved {len(vehicles)} vehicles from server")

# Filter only today's data (25/09/29)
today_vehicles = []
for vehicle in vehicles:
    dt = vehicle.get('DataDateTime', '')
    if '25/09/29' in dt:  # Today's data only
        today_vehicles.append(vehicle)

print(f"Found {len(today_vehicles)} vehicles with today's data (25/09/29)")

if not today_vehicles:
    print("No today's data to send")
    exit(0)

# Convert numeric fields and fix VehicleCD
numeric_fields = [
    "AllStateFontColorIndex", "BranchCD", "CurrentWorkCD", "DataFilterType",
    "DispFlag", "DriverCD", "GPSDirection", "GPSEnable", "GPSLatitude",
    "GPSLongitude", "GPSSatelliteNum", "OperationState", "ReciveEventType",
    "RecivePacketType", "ReciveWorkCD", "Revo", "Speed", "SubDriverCD",
    "TempState"  # Removed VehicleCD from here
]

converted_data = []
vehicle_codes_used = set()

for vehicle in today_vehicles:
    converted_vehicle = vehicle.copy()

    # Convert numeric fields
    for field in numeric_fields:
        if field in converted_vehicle:
            converted_vehicle[field] = convert_to_number(converted_vehicle[field])

    # Generate unique VehicleCD from VehicleName
    vehicle_name = converted_vehicle.get('VehicleName', '')
    vehicle_code = extract_vehicle_code(vehicle_name)

    # Ensure uniqueness by adding offset if needed
    original_code = vehicle_code
    offset = 0
    while vehicle_code in vehicle_codes_used:
        offset += 1
        vehicle_code = original_code + offset

    vehicle_codes_used.add(vehicle_code)
    converted_vehicle['VehicleCD'] = vehicle_code

    # Format datetime field
    if "DataDateTime" in converted_vehicle:
        converted_vehicle["DataDateTime"] = format_datetime(converted_vehicle["DataDateTime"])

    converted_data.append(converted_vehicle)

print(f"Converted {len(converted_data)} today's vehicles with unique VehicleCD values")

# Show sample of today's data being sent
if converted_data:
    print("\nSample of TODAY's data being sent:")
    for i, sample in enumerate(converted_data[:5]):
        print(f"\nVehicle {i+1}:")
        print(f"  VehicleName: {sample.get('VehicleName')}")
        print(f"  VehicleCD: {sample.get('VehicleCD')} (unique)")
        print(f"  DataDateTime: {sample.get('DataDateTime')}")

    # Check for unique VehicleCD values
    unique_codes = set(v['VehicleCD'] for v in converted_data)
    print(f"\nTotal unique VehicleCD values: {len(unique_codes)}")

# Step 2: Send TODAY's data to Hono API
print(f"\nSending {len(converted_data)} vehicles with unique IDs to Hono API...")
response2 = requests.post(
    "https://hono-api.mtamaramu.com/api/dtakologs",
    headers={"Content-Type": "application/json; charset=utf-8"},
    json=converted_data
)

print(f"Response status: {response2.status_code}")
if response2.status_code == 200 or response2.status_code == 201:
    print("Success! Today's data sent with unique VehicleCD values")

    # Step 3: Verify the data was stored
    print("\nVerifying data in Hono API...")
    response3 = requests.get("https://hono-api.mtamaramu.com/api/dtakologs/currentListAll")
    response3.encoding = 'utf-8'

    if response3.status_code == 200:
        all_data = response3.json()
        print(f"Total records now in Hono API: {len(all_data)}")

        # Count today's records
        today_count = 0
        for record in all_data:
            dt = record.get('DataDateTime', '')
            if '2025-09-29' in dt:
                today_count += 1

        print(f"Records with 2025-09-29 (today) in Hono API: {today_count}")

        if len(all_data) > 229:
            print(f"\n✅ Successfully added {len(all_data) - 229} new records!")
else:
    print(f"Error response: {response2.text[:500]}")