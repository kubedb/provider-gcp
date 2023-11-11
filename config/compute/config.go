package compute

import (
	"github.com/upbound/upjet/pkg/config"
	"kubedb.dev/provider-gcp/config/common"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("google_compute_firewall", func(r *config.Resource) {
		r.References["network"] = config.Reference{
			Type:      "Network",
			Extractor: common.PathSelfLinkExtractor,
		}
	})
}
