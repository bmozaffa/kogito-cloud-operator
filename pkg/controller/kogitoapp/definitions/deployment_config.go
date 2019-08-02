package definitions

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultReplicas      = int32(1)
	deploymentConfigKind = "DeploymentConfig"
)

type deploymentConfigResource struct{}

func (d *deploymentConfigResource) New(kogitoApp *v1alpha1.KogitoApp, runnerBC *buildv1.BuildConfig, sa *corev1.ServiceAccount, dockerImage *dockerv10.DockerImage) (dc *appsv1.DeploymentConfig, err error) {
	// should be a struct of dependencies that have constant names
	if err = d.checkDeploymentDependencies(map[string]interface{}{
		"BuildConfig":    runnerBC,
		"ServiceAccount": sa,
		"Image":          dockerImage,
	}); err != nil {
		return dc, err
	}

	dc = &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Spec.Name,
			Namespace: kogitoApp.Namespace,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: kogitoApp.Spec.Name,
							// this conversion will be removed in future versions
							Env: shared.FromEnvToEnvVar(kogitoApp.Spec.Env),
							// this conversion will be removed in future versions
							Resources:       *shared.FromResourcesToResourcesRequirements(kogitoApp.Spec.Resources),
							Image:           runnerBC.Spec.Output.To.Name,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					ServiceAccountName: sa.Name,
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				{Type: appsv1.DeploymentTriggerOnConfigChange},
				{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{kogitoApp.Spec.Name},
						From:           *runnerBC.Spec.Output.To,
					},
				},
			},
		},
	}

	dc.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind(deploymentConfigKind))

	addDefaultMeta(&dc.ObjectMeta, kogitoApp)
	addDefaultMeta(&dc.Spec.Template.ObjectMeta, kogitoApp)
	addDefaultLabels(dc.Spec.Selector, kogitoApp)
	d.setReplicas(kogitoApp, dc)

	return dc, nil
}

// checkDeploymentDependencies sanity check to create the DeploymentConfig properly
func (d *deploymentConfigResource) checkDeploymentDependencies(deps map[string]interface{}) (err error) {
	for dep, obj := range deps {
		if obj == nil {
			err = fmt.Errorf("Impossible to create the DeploymentConfig without a reference to %s", dep)
			break
		}
	}

	return err
}

// setReplicas defines the number of container replicas that this DeploymentConfig will have
func (d *deploymentConfigResource) setReplicas(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	replicas := defaultReplicas
	if kogitoApp.Spec.Replicas != nil {
		replicas = *kogitoApp.Spec.Replicas
	}
	dc.Spec.Replicas = replicas
}

// setPortsAndProbes defines the ports and probes based on the docker image for this DeploymentConfig
func (d *deploymentConfigResource) setPortsAndProbes(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) {

}
