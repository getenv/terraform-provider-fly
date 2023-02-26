resource "fly_certificates" "things" {
  app = "getenv-terraform-provider-fly-test"
  app_id = "getenv-terraform-provider-fly-test"
  host = "test_Host"
}
