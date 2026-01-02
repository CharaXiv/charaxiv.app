output "bucket_name" {
  value = google_storage_bucket.assets.name
}

output "bucket_url" {
  value = google_storage_bucket.assets.url
}

output "service_account_email" {
  value = google_service_account.app.email
}

output "service_account_key" {
  value     = var.create_sa_key ? google_service_account_key.app[0].private_key : null
  sensitive = true
}
