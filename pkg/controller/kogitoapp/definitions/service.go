package definitions

import (
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newService(kogitoApp *v1alpha1.KogitoApp, deploymentConfig *appsv1.DeploymentConfig) (service *corev1.Service, err error) {
	ports, err := buildServicePorts(deploymentConfig)
	if err != nil {
		return service, err
	}

	service = &corev1.Service{
		ObjectMeta: deploymentConfig.ObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: deploymentConfig.Spec.Selector,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}

	setGroupVersionKind(&service.TypeMeta, ServiceKind)
	addDefaultMeta(&service.ObjectMeta, kogitoApp)

	return service, nil
}

func buildServicePorts(deploymentConfig *appsv1.DeploymentConfig) (ports []corev1.ServicePort, err error) {
	// for now, we should have at least 1 container defined.
	if len(deploymentConfig.Spec.Template.Spec.Containers) == 0 ||
		len(deploymentConfig.Spec.Template.Spec.Containers[0].Ports) == 0 {
		return ports,
			fmt.Errorf("The deploymentConfig spec '%s' doesn't have any ports exposed. Impossible to create ServicePorts", deploymentConfig.Name)
	}

	ports = []corev1.ServicePort{}
	for _, port := range deploymentConfig.Spec.Template.Spec.Containers[0].Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.ContainerPort,
			TargetPort: intstr.FromInt(int(port.ContainerPort)),
		},
		)
	}
	return ports, nil
}
