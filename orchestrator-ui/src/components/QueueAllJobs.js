import React, { useState } from "react";
import API_CONFIG from "../config";

function extractQueuedJobIds(payload) {
  if (!payload) return [];

  if (Array.isArray(payload)) {
    return payload
      .map((item) => item?.job_id || item?.JobID || item?.id || item?.ID)
      .filter(Boolean);
  }

  if (typeof payload === "object") {
    const jobsQueued = payload["Jobs Queued"] || payload["jobs_queued"] || payload.jobsQueued;
    if (Array.isArray(jobsQueued)) {
      return jobsQueued
        .map((item) => item?.job_id || item?.JobID || item?.id || item?.ID)
        .filter(Boolean);
    }

    return Object.values(payload)
      .flatMap((value) => {
        if (Array.isArray(value)) {
          return value
            .map((item) => item?.job_id || item?.JobID || item?.id || item?.ID)
            .filter(Boolean);
        }
        return [value]?.filter(Boolean);
      })
      .filter(Boolean);
  }

  return [];
}

function QueueAllJobs() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [jobIds, setJobIds] = useState([]);
  const [error, setError] = useState("");

  const handleQueueAll = async () => {
    setLoading(true);
    setError("");
    setResult(null);
    setJobIds([]);

    try {
      const apiUrl = API_CONFIG.ENDPOINTS.QUEUE_ALL_JOBS;
      const response = await fetch(apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({}),
      });

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data = await response.json();
      setResult(data);
      setJobIds(extractQueuedJobIds(data));
    } catch (err) {
      setError(err.message || "Unable to queue jobs.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <div className="page-header">
        <h3>Queue All Jobs</h3>
      </div>

      {error && <div className="alert-error">{error}</div>}

      <div className="form-section center-actions">
        <button type="button" onClick={handleQueueAll} disabled={loading}>
          {loading ? "Queueing..." : "Queue All Jobs"}
        </button>
      </div>

      {result && (
        <div className="form-section">
          <h4>Queued Job IDs</h4>
          {jobIds.length > 0 ? (
            <ul>
              {jobIds.map((jobId) => (
                <li key={jobId}>{jobId}</li>
              ))}
            </ul>
          ) : (
            <pre>{JSON.stringify(result, null, 2)}</pre>
          )}
        </div>
      )}
    </div>
  );
}

export default QueueAllJobs;
