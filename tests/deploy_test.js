import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    vus: 5,                 // number of parallel users
    duration: "20s",        // run continuously
};

export default function () {

    const payload = JSON.stringify({
        app: "Digitron orchestrator",
        site: "3e5c21bc-2fef-4fd7-a2d0-60fc6b3260ad",
        deploytype: "compose"
    });

    const params = {
        headers: {
            "Content-Type": "application/json"
        }
    };

    // 1️⃣ Start deployment
    const res = http.post("http://localhost:8080/api/v1/deploy", payload, params);

    check(res, {
        "deploy request succeeded": (r) => r.status === 200,
    });

    if (res.status !== 200) {
        return;
    }

    const deployID = res.json("data[0]"); // your output: ["uuid"]
    if (!deployID) return;

    // 2️⃣ Stream deployment in blocking mode (long-lived request)
    const streamUrl = `http://localhost:8080/api/v1/deployments/${deployID}/stream`;
    const streamRes = http.get(streamUrl, { timeout: "60s" });

    check(streamRes, {
        "stream returned 200": (r) => r.status === 200,
    });

    // Optional: log for debug
    console.log(`Stream result for ${deployID}: ${streamRes.body}`);

    sleep(1);
}
