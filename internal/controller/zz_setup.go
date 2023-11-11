/*
Copyright 2021 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/upbound/upjet/pkg/controller"

	firewall "kubedb.dev/provider-gcp/internal/controller/compute/firewall"
	network "kubedb.dev/provider-gcp/internal/controller/compute/network"
	networkpeering "kubedb.dev/provider-gcp/internal/controller/compute/networkpeering"
	providerconfig "kubedb.dev/provider-gcp/internal/controller/providerconfig"
	instance "kubedb.dev/provider-gcp/internal/controller/redis/instance"
	database "kubedb.dev/provider-gcp/internal/controller/spanner/database"
	databaseiammember "kubedb.dev/provider-gcp/internal/controller/spanner/databaseiammember"
	instancespanner "kubedb.dev/provider-gcp/internal/controller/spanner/instance"
	instanceiammember "kubedb.dev/provider-gcp/internal/controller/spanner/instanceiammember"
	databasesql "kubedb.dev/provider-gcp/internal/controller/sql/database"
	databaseinstance "kubedb.dev/provider-gcp/internal/controller/sql/databaseinstance"
	sourcerepresentationinstance "kubedb.dev/provider-gcp/internal/controller/sql/sourcerepresentationinstance"
	sslcert "kubedb.dev/provider-gcp/internal/controller/sql/sslcert"
	user "kubedb.dev/provider-gcp/internal/controller/sql/user"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		firewall.Setup,
		network.Setup,
		networkpeering.Setup,
		providerconfig.Setup,
		instance.Setup,
		database.Setup,
		databaseiammember.Setup,
		instancespanner.Setup,
		instanceiammember.Setup,
		databasesql.Setup,
		databaseinstance.Setup,
		sourcerepresentationinstance.Setup,
		sslcert.Setup,
		user.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
