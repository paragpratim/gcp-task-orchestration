const runtimeConfig = typeof window !== "undefined" ? window.__APP_CONFIG__ || {} : {};
const configuredBaseUrl = runtimeConfig.BASE_URL || process.env.REACT_APP_BACKEND_URL || "";
const baseUrl = configuredBaseUrl.replace(/\/$/, "");

const API_CONFIG = {
  BASE_URL: baseUrl,
  ENDPOINTS: {
    CREATE_JOB: `${baseUrl}/api/v1/intake/job`,
    JOB_STATUS: `${baseUrl}/api/v1/intake/jobs/status`,
    QUEUE_ALL_JOBS: `${baseUrl}/api/v1/intake/jobs/queue`,
  },
};

export default API_CONFIG;