/*
Copyright 2021 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/upbound/upjet/pkg/controller"

	firewall "kubeform.dev/provider-gcp/internal/controller/compute/firewall"
	network "kubeform.dev/provider-gcp/internal/controller/compute/network"
	networkpeering "kubeform.dev/provider-gcp/internal/controller/compute/networkpeering"
	providerconfig "kubeform.dev/provider-gcp/internal/controller/providerconfig"
	instance "kubeform.dev/provider-gcp/internal/controller/redis/instance"
	database "kubeform.dev/provider-gcp/internal/controller/spanner/database"
	databaseiammember "kubeform.dev/provider-gcp/internal/controller/spanner/databaseiammember"
	instancespanner "kubeform.dev/provider-gcp/internal/controller/spanner/instance"
	instanceiammember "kubeform.dev/provider-gcp/internal/controller/spanner/instanceiammember"
	databasesql "kubeform.dev/provider-gcp/internal/controller/sql/database"
	databaseinstance "kubeform.dev/provider-gcp/internal/controller/sql/databaseinstance"
	sourcerepresentationinstance "kubeform.dev/provider-gcp/internal/controller/sql/sourcerepresentationinstance"
	sslcert "kubeform.dev/provider-gcp/internal/controller/sql/sslcert"
	user "kubeform.dev/provider-gcp/internal/controller/sql/user"
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
