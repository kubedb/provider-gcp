/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"
	"kubedb.dev/provider-gcp/config/compute"
	"kubedb.dev/provider-gcp/config/redis"
	"kubedb.dev/provider-gcp/config/spanner"
	"kubedb.dev/provider-gcp/config/sql"

	ujconfig "github.com/crossplane/upjet/pkg/config"
)

const (
	resourcePrefix = "gcp"
	modulePath     = "kubedb.dev/provider-gcp"
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
		ujconfig.WithRootGroup("gcp.kubedb.com"),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	// API group overrides from Terraform import statements
	for _, r := range pc.Resources {
		groupKindOverride(r)
	}

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
