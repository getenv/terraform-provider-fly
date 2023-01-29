package main

import (
	"context"

	"github.com/getenv/terraform-provider-fly/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/getenv/fly",
	})
}
