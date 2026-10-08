resource "google_storage_bucket" "test_buckets" {
  for_each = toset([
    "orchestrator-test-bucket-01-${var.env}-${replace(var.project_id, "-", "")}",
    "orchestrator-test-bucket-02-${var.env}-${replace(var.project_id, "-", "")}",
  ])

  name          = each.value
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  labels = {
    env     = var.env
    purpose = "testing"
  }
}

resource "google_bigquery_dataset" "test_datasets" {
  for_each = toset([
    "orchestrator_test_dataset_01_${var.env}",
    "orchestrator_test_dataset_02_${var.env}",
  ])

  dataset_id = each.value
  location   = var.region

  delete_contents_on_destroy = true

  labels = {
    env     = var.env
    purpose = "testing"
  }
}
