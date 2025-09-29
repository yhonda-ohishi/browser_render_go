// Test script to send vehicle data to Hono API
const fetch = require('node-fetch');

async function testHonoAPI() {
    try {
        // Step 1: Get vehicle data from the server
        console.log("Fetching vehicle data from server...");
        const response1 = await fetch("http://133.18.115.234:8080/v1/vehicle/data", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                branch_id: "",
                filter_id: "0",
                force_login: false
            })
        });

        const data = await response1.json();
        console.log(`Retrieved ${data.data ? data.data.length : 0} vehicles`);

        // Step 2: Send data to Hono API
        console.log("Sending data to Hono API...");
        const response2 = await fetch("https://hono-api.mtamaramu.com/api/dtakologs", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(data["data"])
        });

        console.log("after insert");
        console.log(`Response status: ${response2.status}`);

        const result = await response2.text();
        console.log("Response from Hono API:", result);

    } catch (error) {
        console.error("Error:", error);
    }
}

// Run the test
testHonoAPI();