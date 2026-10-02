variable "project_id" {
  description = "The GCP project ID where resources will be created."
  type        = string
}

variable "region" {
  description = "The GCP region where resources will be created."
  type        = string
}

variable "my_domain" {
  description = "The custom domain to be used for Cloud Run services."
  type        = string
}

variable "env" {
  description = "Deployment environment used to build queue names for the task orchestrator."
  type        = string
  default     = "dev"
}

variable "iap_support_email" {
  description = "Support email used when creating the Google IAP OAuth brand for the project. Leave blank to default to gcp-organization-admins@<my_domain>."
  type        = string
  default     = ""
}

variable "iap_application_title" {
  description = "Application title shown in the IAP OAuth brand UI."
  type        = string
  default     = "Task Orchestrator"
}

