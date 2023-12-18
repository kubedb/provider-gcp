package controller

import (
	"context"
	"sync"

	"github.com/crossplane/upjet/pkg/controller"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	setupFns = map[schema.GroupKind]func(ctrl.Manager, controller.Options) error{
		schema.GroupKind{"compute.gcp.kubedb.com", "Firewall"}:                 firewall.Setup,
		schema.GroupKind{"compute.gcp.kubedb.com", "Network"}:                  network.Setup,
		schema.GroupKind{"compute.gcp.kubedb.com", "NetworkPeering"}:           networkpeering.Setup,
		schema.GroupKind{"providerconfig.gcp.kubedb.com", ""}:                  providerconfig.Setup,
		schema.GroupKind{"redis.gcp.kubedb.com", "Instance"}:                   instance.Setup,
		schema.GroupKind{"spanner.gcp.kubedb.com", "Database"}:                 database.Setup,
		schema.GroupKind{"spanner.gcp.kubedb.com", "DatabaseIAMMember"}:        databaseiammember.Setup,
		schema.GroupKind{"spanner.gcp.kubedb.com", "Instance"}:                 instancespanner.Setup,
		schema.GroupKind{"spanner.gcp.kubedb.com", "InstanceIAMMember"}:        instanceiammember.Setup,
		schema.GroupKind{"sql.gcp.kubedb.com", "Database"}:                     databasesql.Setup,
		schema.GroupKind{"sql.gcp.kubedb.com", "DatabaseInstance"}:             databaseinstance.Setup,
		schema.GroupKind{"sql.gcp.kubedb.com", "SourceRepresentationInstance"}: sourcerepresentationinstance.Setup,
		schema.GroupKind{"sql.gcp.kubedb.com", "SSLCert"}:                      sslcert.Setup,
		schema.GroupKind{"sql.gcp.kubedb.com", "User"}:                         user.Setup,
	}
)

//package controller

var (
	setupDone = map[schema.GroupKind]bool{}
	mu        sync.RWMutex
)

type CustomResourceReconciler struct {
	mgr ctrl.Manager
	o   controller.Options
}

func NewCustomResourceReconciler(mgr ctrl.Manager, o controller.Options) *CustomResourceReconciler {
	return &CustomResourceReconciler{mgr: mgr, o: o}
}

func (r *CustomResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var crd apiextensions.CustomResourceDefinition
	if err := r.mgr.GetClient().Get(ctx, req.NamespacedName, &crd); err != nil {
		log.Error(err, "unable to fetch CustomResourceDefinition")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	gk := schema.GroupKind{
		Group: crd.Spec.Group,
		Kind:  crd.Spec.Names.Kind,
	}
	mu.Lock()
	defer mu.Unlock()
	_, found := setupDone[gk]
	if found {
		return ctrl.Result{}, nil
	}
	setup, found := setupFns[gk]
	if found {
		setup(r.mgr, r.o)
		setupDone[gk] = true
	}

	return ctrl.Result{}, nil
}

func (r *CustomResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensions.CustomResourceDefinition{}).
		Complete(r)
}
