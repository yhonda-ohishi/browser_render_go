import json
import requests
from datetime import datetime

def convert_to_number(value):
    """Convert string to number if possible, otherwise return original value"""
    if value is None or value == "<nil>" or value == "":
        return 0
    if isinstance(value, str):
        try:
            # Try to convert to float
            return float(value)
        except ValueError:
            # If conversion fails, return 0 for empty strings
            return 0
    return value

def format_datetime(dt_string):
    """Format datetime to match what the API expects"""
    if not dt_string or dt_string == "<nil>":
        return None
    # The API expects format like "24/12/29 15:30:45"
    # and will prepend "20" to make "2024/12/29"
    try:
        # Parse the incoming format and convert to expected format
        # Remove "20" prefix if it exists to avoid double prefixing
        if dt_string.startswith("20"):
            dt_string = dt_string[2:]
        return dt_string
    except:
        return dt_string

# Step 1: Get vehicle data
print("Fetching vehicle data...")
response1 = requests.post(
    "http://133.18.115.234:8080/v1/vehicle/data",
    headers={"Content-Type": "application/json"},
    json={"branch_id": "", "filter_id": "0", "force_login": False}
)

if response1.status_code != 200:
    print(f"Error fetching data: {response1.status_code}")
    print(response1.text)
    exit(1)

data = response1.json()
vehicles = data.get('data', [])
print(f"Retrieved {len(vehicles)} vehicles")

if not vehicles:
    print("No vehicles to send")
    exit(0)

# Convert numeric fields from string to number
numeric_fields = [
    "AllStateFontColorIndex", "BranchCD", "CurrentWorkCD", "DataFilterType",
    "DispFlag", "DriverCD", "GPSDirection", "GPSEnable", "GPSLatitude",
    "GPSLongitude", "GPSSatelliteNum", "OperationState", "ReciveEventType",
    "RecivePacketType", "ReciveWorkCD", "Revo", "Speed", "SubDriverCD",
    "TempState", "VehicleCD"
]

converted_data = []
for vehicle in vehicles:
    converted_vehicle = vehicle.copy()

    # Convert numeric fields
    for field in numeric_fields:
        if field in converted_vehicle:
            converted_vehicle[field] = convert_to_number(converted_vehicle[field])

    # Format datetime field
    if "DataDateTime" in converted_vehicle:
        converted_vehicle["DataDateTime"] = format_datetime(converted_vehicle["DataDateTime"])

    converted_data.append(converted_vehicle)

print(f"Converted {len(converted_data)} vehicles")

# Show sample of data being sent
if converted_data:
    print("\nSample vehicle data being sent:")
    sample = converted_data[0]
    print(f"  VehicleName: {sample.get('VehicleName')}")
    print(f"  DataDateTime: {sample.get('DataDateTime')}")
    print(f"  BranchName: {sample.get('BranchName')}")
    print(f"  GPSLatitude: {sample.get('GPSLatitude')} (type: {type(sample.get('GPSLatitude'))})")
    print(f"  Speed: {sample.get('Speed')} (type: {type(sample.get('Speed'))})")

# Step 2: Send to Hono API
print("\nSending to Hono API...")
response2 = requests.post(
    "https://hono-api.mtamaramu.com/api/dtakologs",
    headers={"Content-Type": "application/json; charset=utf-8"},
    json=converted_data
)

print(f"Response status: {response2.status_code}")
if response2.status_code == 200 or response2.status_code == 201:
    print("Success! Data sent to Hono API")

    # Step 3: Verify the data was stored
    print("\nVerifying data in Hono API...")
    response3 = requests.get("https://hono-api.mtamaramu.com/api/dtakologs/currentListAll")
    response3.encoding = 'utf-8'

    if response3.status_code == 200:
        all_data = response3.json()
        print(f"Total records now in Hono API: {len(all_data) if isinstance(all_data, list) else 'Unknown'}")

        if isinstance(all_data, list) and len(all_data) > 0:
            # Show the most recent record
            latest = all_data[-1]
            print("\nLatest record in API:")
            print(f"  VehicleName: {latest.get('VehicleName')}")
            print(f"  DataDateTime: {latest.get('DataDateTime')}")
            print(f"  BranchName: {latest.get('BranchName')}")
else:
    print(f"Error response: {response2.text[:500]}")
    print(f"Response headers: {response2.headers}")