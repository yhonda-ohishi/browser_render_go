import json
import requests

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

# Step 1: Get vehicle data
response1 = requests.post(
    "http://133.18.115.234:8080/v1/vehicle/data",
    headers={"Content-Type": "application/json"},
    json={"branch_id": "", "filter_id": "0", "force_login": False}
)

data = response1.json()
print(f"Retrieved {len(data.get('data', []))} vehicles")

# Convert numeric fields from string to number
numeric_fields = [
    "AllStateFontColorIndex", "BranchCD", "CurrentWorkCD", "DataFilterType",
    "DispFlag", "DriverCD", "GPSDirection", "GPSEnable", "GPSLatitude",
    "GPSLongitude", "GPSSatelliteNum", "OperationState", "ReciveEventType",
    "RecivePacketType", "ReciveWorkCD", "Revo", "Speed", "SubDriverCD",
    "TempState", "VehicleCD"
]

converted_data = []
for vehicle in data["data"]:
    converted_vehicle = vehicle.copy()
    for field in numeric_fields:
        if field in converted_vehicle:
            converted_vehicle[field] = convert_to_number(converted_vehicle[field])
    converted_data.append(converted_vehicle)

print(f"Converted {len(converted_data)} vehicles")

# Step 2: Send to Hono API
response2 = requests.post(
    "https://hono-api.mtamaramu.com/api/dtakologs",
    headers={"Content-Type": "application/json"},
    json=converted_data
)

print("after insert")
print(f"Response status: {response2.status_code}")
if response2.status_code == 200 or response2.status_code == 201:
    print("Success!")
else:
    print(f"Response: {response2.text[:500]}")