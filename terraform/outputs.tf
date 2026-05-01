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
  value       = module.gcp_task_orchestration.service_account_email
}

output "artifact_registry_repository" {
  description = "The Artifact Registry repository name"
  value       = module.gcp_task_orchestration.artifact_registry_repository
}

output "vpc_name" {
  description = "The VPC name"
  value       = module.gcp_task_orchestration.vpc_name
}

output "subnet_name" {
  description = "The subnet name"
  value       = module.gcp_task_orchestration.subnet_name
}

output "subnetwork_ip_cidr_range" {
  description = "The subnet IP CIDR range"
  value       = module.gcp_task_orchestration.subnetwork_ip_cidr_range
}

output "vpc_connector_name" {
  description = "The Serverless VPC Access connector name"
  value       = module.gcp_task_orchestration.vpc_connector_name
}