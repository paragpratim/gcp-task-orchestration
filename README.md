# GCP Task Orchestration

A Go-based, highly efficient, and configurable Google Cloud orchestration platform for performing and orchestrating various tasks across GCP services.

[![Go CI](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/go-ci.yml/badge.svg?branch=main)](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/go-ci.yml)

[![Deploy Terraform Infrastructure](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/deploy-terraform.yml/badge.svg?branch=main)](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/deploy-terraform.yml)

[![Deploy Task Orchestration Service to Cloud Run](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/deploy-task-orchestration-app.yml/badge.svg)](https://github.com/paragpratim/gcp-task-orchestration/actions/workflows/deploy-task-orchestration-app.yml)

## What this repo contains

- `orchestrator-service/` — Go REST API that accepts job requests, writes status to Firestore, and enqueues Cloud Tasks work.
- `orchestrator-ui/` — React dashboard for submitting jobs and monitoring execution.
- `terraform/` — infrastructure as code for GCP resources.
- `docker-compose.yml` — local emulators for Firestore and Cloud Tasks, plus the API and UI.

## Architecture

![Architecture Diagram](docs/architecture.png)

## Features

### Versions

- V1 — GCS to BigQuery data ingestion workflow using Cloud Tasks, Firestore status tracking, and a monitoring UI.

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
