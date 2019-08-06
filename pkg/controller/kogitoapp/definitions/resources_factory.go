package definitions

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ResourcesFactory gather all resources definitions from this package to make it ease for clients to use this Factory as a single point of reference
type ResourcesFactory struct {
	ServiceAccount   func(kogitoApp *v1alpha1.KogitoApp) (serviceAccount corev1.ServiceAccount)
	RoleBinding      *roleBindingResource
	BuildConfig      *buildConfigResource
	DeploymentConfig *deploymentConfigResource
	Service          func(kogitoApp *v1alpha1.KogitoApp, deploymentConfig *appsv1.DeploymentConfig) (service *corev1.Service, err error)
	Route            *routeResource
}

func New() *ResourcesFactory {
	return &ResourcesFactory{
		ServiceAccount:   newServiceAccount,
		Service:          newService,
	}
}