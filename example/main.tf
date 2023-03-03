resource "fly_volumes" "things" {
  app = "getenv-terraform-provider-fly-test"
  name = "TestVolume"
  region = "lax"
  sizegb = 10
}
