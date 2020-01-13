package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/phillbaker/terraform-provider-elasticsearch/es"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: es.Provider,
	})
}
