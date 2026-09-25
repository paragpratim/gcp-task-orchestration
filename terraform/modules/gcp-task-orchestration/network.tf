# VPC with private access to Google APIs
resource "google_compute_network" "orchestrator_vpc" {
  name                    = "orchestrator-vpc"
  auto_create_subnetworks = false
}

# Subnet for orchestrator cloud run services
resource "google_compute_subnetwork" "orchestrator_subnet" {
  name                     = "orchestrator-subnet"
  ip_cidr_range            = "10.0.0.0/24"
  network                  = google_compute_network.orchestrator_vpc.id
  region                   = var.region
  private_ip_google_access = true

  depends_on = [google_compute_network.orchestrator_vpc]
}