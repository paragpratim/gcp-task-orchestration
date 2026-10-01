import React, { useState } from "react";
import API_CONFIG from "../config";

const defaultForm = {
  name: "",
  description: "",
  source: {
    bucket_name: "",
    prefix: "",
    file_pattern: "",
    file_type: "CSV",
  },
  target: {
    project_id: "",
    dataset_id: "",
    table_name: "",
  },
};

function CreateJobDetails() {
  const [form, setForm] = useState(defaultForm);
  const [submitting, setSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [lastCreatedJob, setLastCreatedJob] = useState(null);

  const handleChange = (event) => {
    const { name, value } = event.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  };

  const handleNestedChange = (section, field, value) => {
    setForm((prev) => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value,
      },
    }));
  };

  const validateForm = () => {
    if (!form.name.trim()) {
      return "Job name is required.";
    }
    if (!form.target.project_id.trim()) {
      return "Project ID is required.";
    }
    if (!form.target.dataset_id.trim()) {
      return "Dataset ID is required.";
    }
    if (!form.target.table_name.trim()) {
      return "Table name is required.";
    }
    if (!form.source.file_pattern.trim() && !form.source.bucket_name.trim()) {
      return "At least one source identifier is required.";
    }
    return "";
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    const validationError = validateForm();
    if (validationError) {
      setErrorMessage(validationError);
      setSuccessMessage("");
      return;
    }

    setSubmitting(true);
    setErrorMessage("");

    const payload = {
      id: null,
      name: form.name,
      description: form.description,
      source: {
        bucket_name: form.source.bucket_name || "",
        prefix: form.source.prefix || "",
        file_pattern: form.source.file_pattern || "",
        file_type: form.source.file_type || "CSV",
      },
      target: {
        project_id: form.target.project_id,
        dataset_id: form.target.dataset_id,
        table_name: form.target.table_name,
      },
      metadata: null,
      created_at: null,
      updated_at: null,
    };

    try {
      const apiUrl = API_CONFIG.ENDPOINTS.CREATE_JOB;

      const response = await fetch(apiUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      setLastCreatedJob(payload);
      setSuccessMessage("Job details submitted successfully.");
      setForm(defaultForm);
    } catch (error) {
      setErrorMessage(error.message || "Unable to create the job right now.");
      setSuccessMessage("");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div>
      <div className="page-header">
        <h3>Create Job Details</h3>
      </div>

      {successMessage && <div className="alert-success">{successMessage}</div>}
      {errorMessage && <div className="alert-error">{errorMessage}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-section">
          <div className="form-group">
            <label htmlFor="name">Job Name</label>
            <input
              id="name"
              type="text"
              name="name"
              value={form.name}
              onChange={handleChange}
              placeholder="Enter a job name"
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              name="description"
              value={form.description}
              onChange={handleChange}
              rows="4"
              placeholder="Job description"
            />
          </div>

          <h4>Source</h4>
          <div className="form-group">
            <label htmlFor="bucket_name">Bucket Name</label>
            <input
              id="bucket_name"
              type="text"
              value={form.source.bucket_name}
              onChange={(e) => handleNestedChange("source", "bucket_name", e.target.value)}
              placeholder="my-bucket"
            />
          </div>

          <div className="form-group">
            <label htmlFor="prefix">Prefix</label>
            <input
              id="prefix"
              type="text"
              value={form.source.prefix}
              onChange={(e) => handleNestedChange("source", "prefix", e.target.value)}
              placeholder="incoming/"
            />
          </div>

          <div className="form-group">
            <label htmlFor="file_pattern">File Pattern</label>
            <input
              id="file_pattern"
              type="text"
              value={form.source.file_pattern}
              onChange={(e) => handleNestedChange("source", "file_pattern", e.target.value)}
              placeholder="*.csv"
            />
          </div>

          <div className="form-group">
            <label htmlFor="file_type">File Type</label>
            <select
              id="file_type"
              value={form.source.file_type}
              onChange={(e) => handleNestedChange("source", "file_type", e.target.value)}
            >
              <option value="CSV">CSV</option>
              <option value="AVRO">AVRO</option>
              <option value="NEWLINE_DELIMITED_JSON">NEWLINE_DELIMITED_JSON</option>
              <option value="DATASTORE_BACKUP">DATASTORE_BACKUP</option>
              <option value="GOOGLE_SHEETS">GOOGLE_SHEETS</option>
              <option value="BIGTABLE">BIGTABLE</option>
              <option value="PARQUET">PARQUET</option>
              <option value="ORC">ORC</option>
              <option value="ML_TF_SAVED_MODEL">ML_TF_SAVED_MODEL</option>
              <option value="ML_XGBOOST_BOOSTER">ML_XGBOOST_BOOSTER</option>
              <option value="ICEBERG">ICEBERG</option>
            </select>
          </div>

          <h4>Target</h4>
          <div className="form-group">
            <label htmlFor="project_id">Target Project ID</label>
            <input
              id="project_id"
              type="text"
              value={form.target.project_id}
              onChange={(e) => handleNestedChange("target", "project_id", e.target.value)}
              placeholder="project-id"
            />
          </div>

          <div className="form-group">
            <label htmlFor="dataset_id">Dataset ID</label>
            <input
              id="dataset_id"
              type="text"
              value={form.target.dataset_id}
              onChange={(e) => handleNestedChange("target", "dataset_id", e.target.value)}
              placeholder="analytics_dataset"
            />
          </div>

          <div className="form-group">
            <label htmlFor="table_name">Table Name</label>
            <input
              id="table_name"
              type="text"
              value={form.target.table_name}
              onChange={(e) => handleNestedChange("target", "table_name", e.target.value)}
              placeholder="orders_data"
            />
          </div>

          <button type="submit" disabled={submitting}>
            {submitting ? "Submitting..." : "Submit"}
          </button>
        </div>
      </form>

      {lastCreatedJob && (
        <div className="form-section">
          <h4>Latest Job Payload</h4>
          <pre>{JSON.stringify(lastCreatedJob, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}

export default CreateJobDetails;
