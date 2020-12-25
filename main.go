package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/phillbaker/terraform-provider-elasticsearch/es"
)

// Generate docs for website
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: es.Provider,
	})
}
