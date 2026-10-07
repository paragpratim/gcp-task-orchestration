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
    const statusApiUrl = API_CONFIG.ENDPOINTS.JOB_STATUS;
    const jobsApiUrl = API_CONFIG.ENDPOINTS.JOBS;
    setLoading(true);

    try {
      const [statusResponse, jobsResponse] = await Promise.all([
        fetch(statusApiUrl),
        fetch(jobsApiUrl),
      ]);

      if (!statusResponse.ok || !jobsResponse.ok) {
        throw new Error("Request failed while loading jobs.");
      }

      const statusData = await statusResponse.json();
      const jobsData = await jobsResponse.json();
      const jobsById = new Map((Array.isArray(jobsData) ? jobsData : []).map((job) => [job.id || job.ID, job]));

      const mergedJobs = (Array.isArray(statusData) ? statusData : []).map((job) => {
        const jobId = job.job_id || job.JobID || job.id || job.ID;
        const jobDefinition = jobsById.get(jobId) || {};

        return {
          ...job,
          source: job.source || job.Source || jobDefinition.source || jobDefinition.Source || {},
          target: job.target || job.Target || jobDefinition.target || jobDefinition.Target || {},
        };
      });

      setJobs(mergedJobs);
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
    return jobs.filter((job) => (job.status || job.Status) === activeFilter);
  }, [jobs, activeFilter]);

  const formatSource = (job) => {
    const source = job?.source ?? job?.Source ?? {};
    const bucketName = source.bucket_name || source.bucketName || "";
    const prefix = source.prefix || "";
    const filePattern = source.file_pattern || source.filePattern || "";

    if (!bucketName && !prefix && !filePattern) {
      return "N/A";
    }

    const cleanPrefix = prefix.replace(/^\/+|\/+$/g, "");
    const cleanPattern = filePattern.replace(/^\/+|\/+$/g, "");

    let value = "";
    if (bucketName) {
      value = `gs://${bucketName}`;
    }
    if (cleanPrefix) {
      value = value ? `${value}/${cleanPrefix}` : `/${cleanPrefix}`;
    }
    if (cleanPattern) {
      value = value ? `${value}/${cleanPattern}` : `/${cleanPattern}`;
    }

    return value || "N/A";
  };

  const formatTarget = (job) => {
    const target = job?.target ?? job?.Target ?? {};
    const projectId = target.project_id || target.projectId || "";
    const datasetId = target.dataset_id || target.datasetId || "";
    const tableName = target.table_name || target.tableName || "";

    const targetValue = [projectId, datasetId, tableName].filter(Boolean).join(".");
    return targetValue || "N/A";
  };

  const getJobId = (job) => job?.job_id || job?.JobID || job?.id || job?.ID || "N/A";
  const getStatus = (job) => job?.status || job?.Status || "N/A";
  const getMessage = (job) => job?.message || job?.Message || "N/A";
  const getUpdatedAt = (job) => {
    const updatedAtValue = job?.updated_at || job?.UpdatedAt || job?.updatedAt;
    return updatedAtValue ? new Date(updatedAtValue).toLocaleString() : "N/A";
  };

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

        {filteredJobs.length === 0 ? (
          <div className="empty-state">No jobs found for this filter.</div>
        ) : (
          <div className="monitor-table-wrapper">
            <table className="monitor-table">
              <thead>
                <tr>
                  <th>JobId</th>
                  <th>Source</th>
                  <th>Target</th>
                  <th>Status</th>
                  <th>Message</th>
                  <th>Updated</th>
                </tr>
              </thead>
              <tbody>
                {filteredJobs.map((job) => (
                  <tr key={getJobId(job) + (job?.updated_at || job?.UpdatedAt || "")}>
                    <td>{getJobId(job)}</td>
                    <td>{formatSource(job)}</td>
                    <td>{formatTarget(job)}</td>
                    <td className="status-cell">
                      <span className={`status-badge ${statusClassMap[getStatus(job)] || "status-queued"}`}>
                        {getStatus(job)}
                      </span>
                    </td>
                    <td className="message-cell">{getMessage(job)}</td>
                    <td>{getUpdatedAt(job)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

export default MonitorJobs;
