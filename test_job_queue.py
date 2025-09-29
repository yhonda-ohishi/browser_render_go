import requests
import json
import time

print("Testing Job Queue System...")
print("="*50)

# Test port 8094 with job queue system
base_url = "http://localhost:8094"

# Step 1: Create a job
print("\n1. Creating a job...")
url = f"{base_url}/v1/vehicle/data"
response = requests.get(url)

if response.status_code == 202:
    data = response.json()
    job_id = data.get('job_id')
    print(f"   Job created: {job_id}")
    print(f"   Status: {data.get('status')}")
    print(f"   Message: {data.get('message')}")
else:
    print(f"   Error: {response.status_code} - {response.text}")
    exit(1)

# Step 2: Check job status periodically
print("\n2. Checking job status...")
job_url = f"{base_url}/v1/job/{job_id}"
max_attempts = 60  # 60 seconds max wait
attempt = 0

while attempt < max_attempts:
    time.sleep(1)
    response = requests.get(job_url)

    if response.status_code == 200:
        job_data = response.json()
        status = job_data.get('status')
        print(f"   [{attempt+1}s] Status: {status}", end="")

        if status == 'completed':
            print(" ✓")
            print(f"\n3. Job completed successfully!")
            print(f"   Vehicle Count: {job_data.get('vehicle_count')}")

            if job_data.get('hono_response'):
                hono = job_data['hono_response']
                print(f"\n   Hono API Response:")
                print(f"     Success: {hono.get('success')}")
                print(f"     Records Added: {hono.get('records_added')}")
                print(f"     Total Records: {hono.get('total_records')}")
                print(f"     Message: {hono.get('message')}")
            break

        elif status == 'failed':
            print(" ✗")
            print(f"\n   Job failed: {job_data.get('error')}")
            break
        else:
            print("")  # New line for next status check
    else:
        print(f"\n   Error checking status: {response.status_code}")
        break

    attempt += 1

if attempt >= max_attempts:
    print(f"\n   Timeout waiting for job completion")

# Step 4: List all jobs
print("\n4. Listing all jobs...")
jobs_url = f"{base_url}/v1/jobs"
response = requests.get(jobs_url)

if response.status_code == 200:
    jobs_data = response.json()
    print(f"   Total jobs: {jobs_data.get('count')}")
    for job in jobs_data.get('jobs', [])[:3]:  # Show first 3 jobs
        print(f"   - {job['id']}: {job['status']} (created: {job['created_at'][:19]})")

print("\n" + "="*50)
print("Test complete!")