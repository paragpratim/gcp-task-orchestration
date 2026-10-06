# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "bigquery.googleapis.com",
    "logging.googleapis.com",
    "artifactregistry.googleapis.com",
    "run.googleapis.com",
    "iap.googleapis.com",
    "storage.googleapis.com",
    "cloudtasks.googleapis.com",
    "firestore.googleapis.com",
    "compute.googleapis.com"
  ])

  service            = each.value
  disable_on_destroy = false
}

# Create service account for task orchestrator application
resource "google_service_account" "orchestrator_service_account" {
  account_id   = "svc-cr-task-orchestrator"
  display_name = "Task Orchestrator Service Account"
  description  = "Service account for task orchestrator application"
}

# Assign multiple roles to service account
resource "google_project_iam_member" "orchestrator_service_account_roles" {
  for_each = toset([
    "roles/bigquery.admin",
    "roles/storage.admin",
    "roles/artifactregistry.writer",
    "roles/cloudtasks.admin",
    # "roles/iam.serviceAccountTokenCreator",
    # "roles/iam.serviceAccountUser",
    "roles/datastore.user",
    # "roles/iap.httpsResourceAccessor",
    "roles/run.admin",
    # "roles/run.invoker"
  ])

  project    = var.project_id
  role       = each.value
  member     = "serviceAccount:${google_service_account.orchestrator_service_account.email}"
  depends_on = [google_service_account.orchestrator_service_account]
}

# Create Artifact Registry repository
resource "google_artifact_registry_repository" "task_orchestrator_repository" {
  location      = var.region
  repository_id = "task-orchestrator"
  description   = "Docker repository for task orchestrator applications"
  format        = "DOCKER"

  depends_on = [google_project_service.required_apis]
}

# Grant IAP access to domain
resource "google_project_iam_member" "iap_domain_access" {
  project = var.project_id
  role    = "roles/iap.httpsResourceAccessor"
  member  = "domain:${var.my_domain}"

  depends_on = [google_project_service.required_apis]
}

# Create Firestore database for state management
resource "google_firestore_database" "orchestrator_firestore" {
  name             = "orchestrator"
  project          = var.project_id
  location_id      = var.region
  type             = "FIRESTORE_NATIVE"
  database_edition = "STANDARD"

  depends_on = [google_project_service.required_apis]
}