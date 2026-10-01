import React, { useEffect, useMemo, useState } from "react";
import API_CONFIG from "../config";

const statusClassMap = {
  QUEUED: "status-queued",
  SUCCESS: "status-succeeded",
  COMPLETED_SKIPPED: "status-succeeded",
  PROCESSING_GCS: "status-running",
  MOVING_GCS: "status-running",
  COMPLETED_GCS: "status-succeeded",
  FAILED_GCS: "status-failed",
  PROCESSING_BIGQUERY: "status-running",
  COMPLETED_BIGQUERY: "status-succeeded",
  FAILED_BIGQUERY: "status-failed",
  PROCESSING_DATAFLOW: "status-running",
  COMPLETED_DATAFLOW: "status-succeeded",
  FAILED_DATAFLOW: "status-failed",
};

function MonitorJobs() {
  const [jobs, setJobs] = useState([]);
  const [activeFilter, setActiveFilter] = useState("All");
  const [loading, setLoading] = useState(false);

  const loadJobs = async () => {
    const apiUrl = API_CONFIG.ENDPOINTS.JOB_STATUS;
    setLoading(true);

    try {
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data = await response.json();
      setJobs(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Unable to fetch jobs:", error);
      setJobs([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadJobs();
  }, []);

  const statusFilters = [
    "All",
    "QUEUED",
    "PROCESSING_GCS",
    "MOVING_GCS",
    "COMPLETED_GCS",
    "FAILED_GCS",
    "PROCESSING_BIGQUERY",
    "COMPLETED_BIGQUERY",
    "FAILED_BIGQUERY",
    "PROCESSING_DATAFLOW",
    "COMPLETED_DATAFLOW",
    "FAILED_DATAFLOW",
    "SUCCESS",
    "COMPLETED_SKIPPED",
  ];

  const filteredJobs = useMemo(() => {
    if (activeFilter === "All") {
      return jobs;
    }
    return jobs.filter((job) => job.status === activeFilter);
  }, [jobs, activeFilter]);

  const refreshJobs = async () => {
    await loadJobs();
  };

  return (
    <div>
      <div className="page-header monitor-header">
        <h3>Monitor Jobs</h3>
        <button type="button" onClick={refreshJobs} disabled={loading}>
          {loading ? "Refreshing..." : "Refresh"}
        </button>
      </div>

      <div className="form-section">
        <div className="filter-row">
          {statusFilters.map((status) => (
            <button
              key={status}
              type="button"
              className={activeFilter === status ? "filter-chip active" : "filter-chip"}
              onClick={() => setActiveFilter(status)}
            >
              {status}
            </button>
          ))}
        </div>

        <div className="monitor-grid">
          {filteredJobs.length === 0 ? (
            <div className="empty-state">No jobs found for this filter.</div>
          ) : (
            filteredJobs.map((job) => (
              <div key={job.job_id || job.JobID || Math.random()} className="job-card">
                <div className="job-card-top">
                  <div>
                    <div className="job-id">{job.job_id || job.JobID}</div>
                    <h4>{job.message || "Job status"}</h4>
                  </div>
                  <span className={`status-badge ${statusClassMap[job.status] || "status-queued"}`}>
                    {job.status}
                  </span>
                </div>

                <div className="job-meta">
                  <span>Updated: {job.updated_at ? new Date(job.updated_at).toLocaleString() : "N/A"}</span>
                  {job.metadata && Object.keys(job.metadata).length > 0 && (
                    <span>Metadata: {JSON.stringify(job.metadata)}</span>
                  )}
                </div>

                {job.message && (
                  <div className="progress-block">
                    <div className="progress-header">
                      <span>Message</span>
                    </div>
                    <div className="message-box">{job.message}</div>
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

export default MonitorJobs;
