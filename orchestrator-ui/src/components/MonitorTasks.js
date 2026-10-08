import React, { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
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

function MonitorTasks() {
  const [searchParams] = useSearchParams();
  const [tasks, setTasks] = useState([]);
  const [jobIdFilter, setJobIdFilter] = useState(searchParams.get("jobId") || "");
  const [loading, setLoading] = useState(false);

  const loadTasks = async () => {
    const apiUrl = API_CONFIG.ENDPOINTS.JOB_STATUS_LOGS;
    setLoading(true);

    try {
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data = await response.json();
      setTasks(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Unable to fetch task logs:", error);
      setTasks([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const paramJobId = searchParams.get("jobId") || "";
    setJobIdFilter(paramJobId);
    loadTasks();
  }, [searchParams]);

  const filteredTasks = useMemo(() => {
    const filterValue = jobIdFilter.trim().toLowerCase();
    const sortedTasks = [...tasks].sort((a, b) => {
      const aTime = new Date(a?.updated_at || a?.UpdatedAt || a?.updatedAt || 0).getTime();
      const bTime = new Date(b?.updated_at || b?.UpdatedAt || b?.updatedAt || 0).getTime();
      return bTime - aTime;
    });

    if (!filterValue) {
      return sortedTasks;
    }

    return sortedTasks.filter((task) => {
      const jobId = task?.job_id || task?.JobID || task?.jobId || "";
      return String(jobId).toLowerCase().includes(filterValue);
    });
  }, [tasks, jobIdFilter]);

  const getStatus = (task) => task?.status || task?.Status || "N/A";
  const getUpdatedAt = (task) => {
    const updatedAtValue = task?.updated_at || task?.UpdatedAt || task?.updatedAt;
    return updatedAtValue ? new Date(updatedAtValue).toLocaleString() : "N/A";
  };

  const refreshTasks = async () => {
    await loadTasks();
  };

  return (
    <div>
      <div className="page-header monitor-header">
        <h3>Monitor Tasks</h3>
        <button type="button" onClick={refreshTasks} disabled={loading}>
          {loading ? "Refreshing..." : "Refresh"}
        </button>
      </div>

      <div className="form-section">
        <div className="filter-row">
          <input
            type="text"
            value={jobIdFilter}
            onChange={(event) => setJobIdFilter(event.target.value)}
            placeholder="Filter by JobId"
            style={{ minWidth: "240px" }}
          />
        </div>

        {filteredTasks.length === 0 ? (
          <div className="empty-state">No task logs found for this JobId filter.</div>
        ) : (
          <div className="monitor-table-wrapper">
            <table className="monitor-table">
              <thead>
                <tr>
                  <th>JobId</th>
                  <th>Message</th>
                  <th>Status</th>
                  <th>TaskId</th>
                  <th>Updated</th>
                </tr>
              </thead>
              <tbody>
                {filteredTasks.map((task, index) => (
                  <tr key={(task?.job_id || task?.JobID || task?.jobId || "task") + (task?.task_id || task?.TaskID || "") + index}>
                    <td>{task?.job_id || task?.JobID || task?.jobId || "N/A"}</td>
                    <td className="message-cell">{task?.message || task?.Message || "N/A"}</td>
                    <td className="status-cell">
                      <span className={`status-badge ${statusClassMap[getStatus(task)] || "status-queued"}`}>
                        {getStatus(task)}
                      </span>
                    </td>
                    <td>{task?.task_id || task?.TaskID || "N/A"}</td>
                    <td>{getUpdatedAt(task)}</td>
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

export default MonitorTasks;
