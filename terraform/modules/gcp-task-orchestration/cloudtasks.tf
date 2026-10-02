# Create service account for cloud tasks
resource "google_service_account" "dispatcher_service_account" {
  account_id   = "svc-cr-task-dispatcher"
  display_name = "Task Dispatcher Service Account"
  description  = "Service account for task dispatcher application"
}

# Create Cloud Tasks queues used by the orchestrator
resource "google_cloud_tasks_queue" "queues" {
  for_each = toset([
    "orchestrator-intake-queue-${var.env}",
    "orchestrator-gcs-queue-${var.env}",
    "orchestrator-bq-queue-${var.env}",
  ])

  name     = each.value
  location = var.region
  project  = var.project_id

  retry_config {
    max_attempts       = 5
    max_retry_duration = "3600s"
    min_backoff        = "5s"
    max_backoff        = "3600s"
    max_doublings      = 3
  }

  rate_limits {
    max_dispatches_per_second = 10
    max_concurrent_dispatches = 3
  }

  depends_on = [google_project_service.required_apis]
}

# Assign project-scoped roles to the dispatcher service account.
# Cloud Run invocation permissions must be granted on the specific Cloud Run service
# resource (for example via google_cloud_run_service_iam_member), not at the project level.
resource "google_project_iam_member" "dispatcher_service_account_roles" {
  for_each = toset([
    "roles/cloudtasks.admin",
    "roles/iap.httpsResourceAccessor",
  ])

  project    = var.project_id
  role       = each.value
  member     = "serviceAccount:${google_service_account.dispatcher_service_account.email}"
  depends_on = [google_service_account.dispatcher_service_account]
}

resource "google_service_account_iam_member" "orchestrator_can_impersonate_dispatcher" {
  for_each = toset([
    "roles/iam.serviceAccountUser",
    "roles/iam.serviceAccountTokenCreator",
  ])

  service_account_id = google_service_account.dispatcher_service_account.name
  role               = each.value
  member             = "serviceAccount:${google_service_account.orchestrator_service_account.email}"

  depends_on = [
    google_service_account.dispatcher_service_account,
    google_service_account.orchestrator_service_account,
  ]
}

data "google_project" "project" {}

# Give the Cloud Tasks background engine permission to use your dispatcher account
resource "google_service_account_iam_member" "cloud_tasks_can_impersonate_dispatcher" {
  service_account_id = google_service_account.dispatcher_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-cloudtasks.iam.gserviceaccount.com"
}

