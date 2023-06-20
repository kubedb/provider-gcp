/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"
	"kubeform.dev/provider-gcp/config/compute"
	"kubeform.dev/provider-gcp/config/redis"
	"kubeform.dev/provider-gcp/config/spanner"
	"kubeform.dev/provider-gcp/config/sql"

	ujconfig "github.com/upbound/upjet/pkg/config"
)

const (
	resourcePrefix = "gcp"
	modulePath     = "kubeform.dev/provider-gcp"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithRootGroup("gcp.kubeform.com"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		compute.Configure,
		sql.Configure,
		spanner.Configure,
		redis.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
