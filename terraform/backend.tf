terraform {
  backend "gcs" {
    bucket = "gcp-task-orchestration-tf-state-747c"
    prefix = "task-orchestration/terraform/state"
  }
}
