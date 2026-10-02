resource "google_iap_brand" "project_brand" {
  project           = var.project_id
  support_email     = var.iap_support_email != "" ? var.iap_support_email : "gcp-organization-admins@${var.my_domain}"
  application_title = var.iap_application_title

  depends_on = [google_project_service.required_apis]
}

resource "google_iap_client" "project_oauth_client" {
  display_name = "${var.iap_application_title} OAuth Client"
  brand        = google_iap_brand.project_brand.name

  depends_on = [google_iap_brand.project_brand]
}
