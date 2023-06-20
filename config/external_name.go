/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/upbound/upjet/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// to work with google network peering you need firewall crd
	"google_compute_network_peering": config.TemplatedStringAsIdentifier("name", "{{ .setup.configuration.project }}/{{ .parameters.network }}/{{ .external_name }}"),
	"google_compute_firewall":        config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/global/firewalls/{{ .external_name }}"),
	// Imported by using the following format: projects/{{project}}/global/networks/{{name}}
	"google_compute_network": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/global/networks/{{ .external_name }}"),
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
