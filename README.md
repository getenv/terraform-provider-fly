# Terraform Fly Provider

This is our very own Terraform provider for [Fly](https://fly.io). The
[official provider](https://github.com/fly-apps/terraform-provider-fly) can't
and won't do what we need, so we wrote our own for now.

## Release

Create a git tag with the `vx.x.x` convention and push it up, just bumping the
PATCH version for now since using this is wildly discouraged for anyone but us.
It would live in our private provider registry if/when we can figure out a way
to do that. 

### TODO

* private TF provider registry
* app management
* ip management
* machine management
