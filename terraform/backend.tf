terraform {
  backend "gcs" {
    bucket = "iceberg-gcp-tf-state-9f85"
    prefix = "task-orchestration/terraform/state"
  }
}
