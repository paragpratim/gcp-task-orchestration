variable "project_id" {
  description = "The GCP project ID"
  type        = string
  default     = ""
}

variable "region" {
  description = "The GCP region"
  type        = string
  default     = ""
}

variable "my_domain" {
  description = "The domain to grant IAP access"
  type        = string
  default     = ""
}

variable "env" {
  description = "The deployment environment used for queue naming"
  type        = string
  default     = "dev"
}

variable "iap_support_email" {
  description = "Support email used when creating the Google IAP OAuth brand for the project. Defaults to gcp-organization-admins@<my_domain> when empty."
  type        = string
  default     = ""
}

variable "iap_application_title" {
  description = "Application title shown in the IAP OAuth brand UI."
  type        = string
  default     = "Task Orchestrator"
}

