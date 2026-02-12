package main

import (
	"context"
	"log"

	// Packages
	providerserver "github.com/hashicorp/terraform-plugin-framework/providerserver"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	providerAddress = "registry.terraform.io/mutablelogic/kaiak"
)

///////////////////////////////////////////////////////////////////////////////
// MAIN

func main() {
	if err := providerserver.Serve(context.Background(), New, providerserver.ServeOpts{
		Address: providerAddress,
	}); err != nil {
		log.Fatal(err)
	}
}
