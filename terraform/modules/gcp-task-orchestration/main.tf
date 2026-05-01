# Create service account
resource "google_service_account" "orchestrator_service_account" {
  account_id   = "svc-cr-task-orchestrator"
  display_name = "Task Orchestrator Service Account"
  description  = "Service account for task orchestrator application"
}

# Enable required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "bigquery.googleapis.com",
    "logging.googleapis.com",
    "artifactregistry.googleapis.com",
    "run.googleapis.com",
    "iap.googleapis.com",
    "storage.googleapis.com"
  ])

  service            = each.value
  disable_on_destroy = false
}

# Assign multiple roles to service account
resource "google_project_iam_member" "orchestrator_service_account_roles" {
  for_each = toset([
    "roles/bigquery.admin",
    "roles/storage.admin",
    "roles/artifactregistry.writer"
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

# VPC with private access to Google APIs
resource "google_compute_network" "orchestrator_vpc" {
  name                    = "orchestrator-vpc"
  auto_create_subnetworks = false
}

# Subnet for orchestrator cloud run services
resource "google_compute_subnetwork" "orchestrator_subnet" {
  name                     = "orchestrator-subnet"
  ip_cidr_range            = "10.0.0.0/24"
  network                  = google_compute_network.orchestrator_vpc.id
  region                   = var.region
  private_ip_google_access = true

  depends_on = [google_compute_network.orchestrator_vpc]
}
