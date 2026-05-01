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

