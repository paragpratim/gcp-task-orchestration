output "project_id" {
  description = "The GCP project ID"
  value       = var.project_id
}

output "region" {
  description = "The GCP region"
  value       = var.region
}

output "service_account_email" {
  description = "The service account email"
  value       = google_service_account.orchestrator_service_account.email
}

output "artifact_registry_repository" {
  description = "The Artifact Registry repository name"
  value       = google_artifact_registry_repository.task_orchestrator_repository.name
}

output "vpc_name" {
  description = "The VPC name"
  value       = google_compute_network.orchestrator_vpc.name
}

output "subnet_name" {
  description = "The subnet name"
  value       = google_compute_subnetwork.orchestrator_subnet.name
}

output "subnetwork_ip_cidr_range" {
  description = "The subnet IP CIDR range"
  value       = google_compute_subnetwork.orchestrator_subnet.ip_cidr_range
}

output "vpc_connector_name" {
  description = "The Serverless VPC Access connector name"
  value       = google_vpc_access_connector.orchestrator_vpc_connector.name
}
