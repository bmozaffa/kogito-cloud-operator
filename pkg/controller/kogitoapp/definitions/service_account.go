package definitions

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ServiceAccountName default name for the SA responsible for running on Kogito App
	ServiceAccountName = "kogito-service"
)

// ServiceAccountDefinition is the factory for new Service Resources
type serviceAccountResource struct {
}

// New creates a new ServiceAccount resource for Kogito App
func (*serviceAccountResource) New(kogitoApp *v1alpha1.KogitoApp) (serviceAccount corev1.ServiceAccount, err error) {
	serviceAccount = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName,
			Namespace: kogitoApp.Namespace,
		},
	}
	addDefaultMeta(&serviceAccount.ObjectMeta, kogitoApp.Name)
	return serviceAccount, nil
}
