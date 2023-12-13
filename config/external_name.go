/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/crossplane/upjet/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// to work with google network peering you need firewall crd
	"google_compute_network_peering": config.TemplatedStringAsIdentifier("name", "{{ .setup.configuration.project }}/{{ .parameters.network }}/{{ .external_name }}"),
	"google_compute_firewall":        config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/global/firewalls/{{ .external_name }}"),
	// Imported by using the following format: projects/{{project}}/global/networks/{{name}}
	"google_compute_network": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/global/networks/{{ .external_name }}"),

	//sql

	// Imported by using the following format: projects/{{project}}/instances/{{name}}
	"google_sql_database_instance": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/instances/{{ .external_name }}"),
	// Imported by using the following format: projects/{{project}}/instances/{{instance}}/databases/{{name}}
	"google_sql_database": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/instances/{{ .parameters.instance }}/databases/{{ .external_name }}"),
	// Imported by using the following format: projects/{{project}}/instances/{{name}}
	"google_sql_source_representation_instance": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/instances/{{ .external_name }}"),
	// Imported by using the following format: my-project/main-instance/me
	"google_sql_user": config.TemplatedStringAsIdentifier("name", "{{ .setup.configuration.project }}/{{ .parameters.instance }}/{{ .external_name }}"),
	// No import
	"google_sql_ssl_cert": config.IdentifierFromProvider,

	// spanner
	//
	// google_spanner_database.default {{instance}}/{{name}}
	"google_spanner_database": config.TemplatedStringAsIdentifier("name", "{{ .parameters.instance }}/instance/{{ .external_name }}"),
	// google_spanner_instance.default {{project}}/{{name}}
	"google_spanner_instance": config.TemplatedStringAsIdentifier("name", "{{ .terraformProviderConfig.project }}/instance/{{ .external_name }}"),
	// google_spanner_instance_iam_member.instance "project-name/instance-name roles/viewer user:foo@example.com"
	"google_spanner_instance_iam_member": config.IdentifierFromProvider,
	// google_spanner_database_iam_member.database "project-name/instance-name/database-name roles/viewer user:foo@example.com"
	"google_spanner_database_iam_member": config.IdentifierFromProvider,

	// redis
	//
	// Imported by using the following format: projects/{{project}}/locations/{{region}}/instances/{{name}}
	"google_redis_instance": config.TemplatedStringAsIdentifier("name", "projects/{{ .setup.configuration.project }}/locations/{{ .parameters.region }}/instances/{{ .external_name }}"),
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
