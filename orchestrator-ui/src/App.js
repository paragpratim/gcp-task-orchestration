import React from "react";
import { BrowserRouter as Router, Routes, Route, Link, Navigate } from "react-router-dom";
import CreateJobDetails from "./components/CreateJobDetails";
import MonitorJobs from "./components/MonitorJobs";
import QueueAllJobs from "./components/QueueAllJobs";
import "./styles.css";

function App() {
  return (
    <Router>
      <div className="container">
        <div className="logo">
          <img src="https://cdn-icons-png.flaticon.com/512/5968/5968705.png" alt="Logo" />
        </div>
        <h1>Task Orchestration</h1>
        <nav style={{ marginBottom: "32px" }}>
          <Link to="/create-job" className="ui-button ui-widget ui-corner-all">Create Job Details</Link>
          <Link to="/monitor-jobs" className="ui-button ui-widget ui-corner-all">Monitor Jobs</Link>
          <Link to="/queue-all-jobs" className="ui-button ui-widget ui-corner-all">Queue All Jobs</Link>
        </nav>
        <Routes>
          <Route path="/" element={<Navigate to="/create-job" replace />} />
          <Route path="/create-job" element={<CreateJobDetails />} />
          <Route path="/monitor-jobs" element={<MonitorJobs />} />
          <Route path="/queue-all-jobs" element={<QueueAllJobs />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;