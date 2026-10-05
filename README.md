# GCP Task Orchestration

A compact Google Cloud workflow for creating, queueing, and monitoring background jobs across GCS, BigQuery, and Dataflow.

## What this repo contains

- `orchestrator-service/` — Go REST API that accepts job requests, writes status to Firestore, and enqueues Cloud Tasks work.
- `orchestrator-ui/` — React dashboard for submitting jobs and monitoring execution.
- `terraform/` — infrastructure as code for GCP resources.
- `docker-compose.yml` — local emulators for Firestore and Cloud Tasks, plus the API and UI.

## Architecture

![Architecture Diagram](docs/architecture.png)

## Quick start

```bash
task up
```

This builds the app and starts the local stack.

- API Swagger UI: http://localhost:8080/swagger/swagger-ui/index.html
- UI: http://localhost:3000

Useful commands:

```bash
task down
task test
task build
task swagger
```

## Repo layout

```text
.
├── docker-compose.yml
├── Taskfile.yml
├── orchestrator-service/
├── orchestrator-ui/
├── terraform/
└── README.md
```

## Notes

- Local development uses emulators for Firestore and Cloud Tasks.
- Production deployment is handled through Terraform and the Go service configuration.
- The service includes Swagger docs under `orchestrator-service/docs`.
