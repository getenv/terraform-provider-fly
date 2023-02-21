resource "fly_secrets" "things" {
  app = "getenv-terraform-provider-fly-test"

  secrets = {
    "THING" = "1"
  }
}
