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

output "firestore_database_name" {
  description = "The Firestore database name created by the orchestrator module"
  value       = module.gcp_task_orchestration.firestore_database_name
}

output "task_queue_names" {
  description = "The Cloud Tasks queue names created by the orchestrator module"
  value       = module.gcp_task_orchestration.task_queue_names
}
output "iap_brand_name" {
  description = "The project-level IAP brand used for the OAuth configuration."
  value       = module.gcp_task_orchestration.iap_brand_name
}

output "iap_oauth_client_id" {
  description = "The OAuth client ID for the project IAP brand."
  value       = module.gcp_task_orchestration.iap_oauth_client_id
}

output "iap_oauth_client_secret" {
  description = "The OAuth client secret for the project IAP brand."
  value       = module.gcp_task_orchestration.iap_oauth_client_secret
  sensitive   = true
}
output "test_bucket_names" {
  description = "Names of the test GCS buckets created for validation"
  value       = values(google_storage_bucket.test_buckets)[*].name
}

output "test_dataset_ids" {
  description = "Dataset IDs of the test BigQuery datasets created for validation"
  value       = values(google_bigquery_dataset.test_datasets)[*].dataset_id
}