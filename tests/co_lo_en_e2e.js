import http from "k6/http";
import { check, sleep } from "k6";
import { Trend } from "k6/metrics";

// Custom metrics
const deployStartTime = new Trend("deploy_start_latency");
const streamDuration = new Trend("deploy_stream_duration");
const deploySuccess = new Trend("deploy_success_ratio");



export const options = {
    scenarios: {
        co_lo_en: {
            executor: "constant-vus",
            vus: 5,               // simulate 5 parallel deployments
            duration: "30s",      // run for 30s
        },
    },
};

// Helper: block until stream ends
function streamDeployment(url, deployID) {
    let start = Date.now();
    let res = http.get(url, { timeout: "120s" });

    let elapsed = Date.now() - start;
    streamDuration.add(elapsed);

    console.log(`Stream completed for ${deployID} (${elapsed} ms)`);

    return res;
}



export default function () {

    // 1️⃣ Deployment request
	const target = {
		site_id: "3e5c21bc-2fef-4fd7-a2d0-60fc6b3260ad",
		host_ids: ["host1", "host2"],
	};
    let body = JSON.stringify({
        app_name: "Digitron orchestrator",
		sites: [target],
        deploy_type: "compose"
    });

    let start = Date.now();
    let res = http.post("http://localhost:8080/api/v1/deployments", body, {
        headers: { "Content-Type": "application/json" }
    });

    // Mark latency
    deployStartTime.add(Date.now() - start);

    check(res, {
        "deploy request OK (200)": (r) => r.status === 200,
    });

    if (res.status !== 200) return;

	//     // Parse JSON response
    // let data = res.json();
	// console.log("Response JSON:", JSON.stringify(data));
	let data = res.json();

	// Check if deployment_ids exists and has at least one element
	let deployID;
	if (data.deployment_ids && data.deployment_ids.length > 0) {
		deployID = data.deployment_ids[0];
		console.log("Deployment ID:", deployID);
	} else {
		console.error("Failed to extract deployment ID: no deployment_ids in response");
	}

    // // Extract deployment ID
    // const deployID = res.json("data[0]");
    // if (!deployID) {
    //     console.error("Failed to extract deployment ID");
    //     return;
    // }

    console.log(`🚀 Deploy started: ${deployID}`);

    // 2️⃣ Stream deployment progress
    const streamUrl = `http://localhost:8080/api/v1/deployments/${deployID}/stream`;

    const streamRes = streamDeployment(streamUrl, deployID);

    check(streamRes, {
        "stream returned 200": (r) => r.status === 200,
    });

    // Log output (optional)
    console.log(`📡 Stream response: ${streamRes.body}`);

    // 3️⃣ Determine success/failure
    const isSuccess = streamRes.body.includes("finished") && !streamRes.body.includes("failed");

    deploySuccess.add(isSuccess ? 1 : 0);

    sleep(1);
}
