terraform {
  backend "gcs" {
    bucket = "task-orchestration-tf-state-381e"
    prefix = "task-orchestration/terraform/state"
  }
}
