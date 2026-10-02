module "gcp_task_orchestration" {
  source = "./modules/gcp-task-orchestration"

  project_id            = var.project_id
  region                = var.region
  my_domain             = var.my_domain
  env                   = var.env
  iap_support_email     = var.iap_support_email
  iap_application_title = var.iap_application_title
}
