package main

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-rok-eks/eks"
)

func main() {
	providerserver.Serve(context.Background(), eks.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/immerok/rok-eks",
	})
}
