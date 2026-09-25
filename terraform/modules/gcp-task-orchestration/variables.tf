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

