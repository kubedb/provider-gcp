package config

import ujconfig "github.com/upbound/upjet/pkg/config"

var (
	resourceGroup = map[string]string{
		"google_compute_firewall":                   "compute",
		"google_compute_network_peering":            "compute",
		"google_compute_network":                    "compute",
		"google_sql_database_instance":              "sql",
		"google_sql_database":                       "sql",
		"google_sql_source_representation_instance": "sql",
		"google_sql_user":                           "sql",
		"google_sql_ssl_cert":                       "sql",
		"google_spanner_database":                   "spanner",
		"google_spanner_instance":                   "spanner",
		"google_spanner_instance_iam_member":        "spanner",
		"google_spanner_database_iam_member":        "spanner",
		"google_redis_instance":                     "redis",
	}
	resourceKind = map[string]string{
		"google_compute_firewall":                   "Firewall",
		"google_compute_network_peering":            "NetworkPeering",
		"google_compute_network":                    "Network",
		"google_sql_database_instance":              "DatabaseInstance",
		"google_sql_database":                       "Database",
		"google_sql_source_representation_instance": "SourceRepresentationInstance",
		"google_sql_user":                           "User",
		"google_sql_ssl_cert":                       "SSLCert",
		"google_spanner_database":                   "Database",
		"google_spanner_instance":                   "Instance",
		"google_spanner_instance_iam_member":        "InstanceIAMMember",
		"google_spanner_database_iam_member":        "DatabaseIAMMember",
		"google_redis_instance":                     "Instance",
	}
)

// default api-group & kind configuration for all resources
func groupKindOverride(r *ujconfig.Resource) {
	if _, ok := resourceGroup[r.Name]; ok {
		r.ShortGroup = resourceGroup[r.Name]
	}

	if _, ok := resourceKind[r.Name]; ok {
		r.Kind = resourceKind[r.Name]
	}
}
