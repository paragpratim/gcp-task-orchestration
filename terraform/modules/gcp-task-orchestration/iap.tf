# NOTE:
# Google IAP OAuth brand/client creation via Terraform is deprecated and often fails
# with a 400 invalid argument in CI or on newly created projects.
#
# Create the brand manually in the GCP Console or via the gcloud IAP flow, then enable
# Cloud Run IAP with the project brand and choose the Custom OAuth / specific control
# option in the service settings.
#
# This avoids breaking infrastructure deploys while keeping the project ready for the
# manual IAP branding step that is required before you can use the OAuth client in
# your Go app.
