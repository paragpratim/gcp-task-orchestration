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

output "firestore_database_name" {
  description = "The Firestore database name"
  value       = google_firestore_database.orchestrator_firestore.name
}

output "task_queue_names" {
  description = "The Cloud Tasks queue names created for the orchestrator"
  value       = values(google_cloud_tasks_queue.queues)[*].name
}
