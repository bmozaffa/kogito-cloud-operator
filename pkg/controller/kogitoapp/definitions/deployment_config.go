package definitions

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultReplicas         = int32(1)
	labelNamespaceSep       = "/"
	orgKieNamespaceLabelKey = "org.kie" + labelNamespaceSep
	labelExposeServices     = "io.openshift.expose-services"
	dockerLabelServicesSep  = ","
	portSep                 = ":"
	portFormatWrongMessage  = "Service %s on " + labelExposeServices + " label in wrong format. Won't be possible to expose Services for this application. Should be PORT_NUMBER:PROTOCOL. e.g. 8080:http"
	defaultExportedProtocol = "http"
)

var defaultProbe = &corev1.Probe{
	TimeoutSeconds:   int32(1),
	PeriodSeconds:    int32(10),
	SuccessThreshold: int32(1),
	FailureThreshold: int32(3),
}

type deploymentConfigResource struct{}

func (d *deploymentConfigResource) New(kogitoApp *v1alpha1.KogitoApp, runnerBC *buildv1.BuildConfig, sa *corev1.ServiceAccount, dockerImage *dockerv10.DockerImage) (dc *appsv1.DeploymentConfig, err error) {
	if err = d.checkDeploymentDependencies(runnerBC, sa); err != nil {
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

	setGroupVersionKind(&dc.TypeMeta, DeploymentConfigKind)
	addDefaultMeta(&dc.ObjectMeta, kogitoApp)
	addDefaultMeta(&dc.Spec.Template.ObjectMeta, kogitoApp)
	addDefaultLabels(&dc.Spec.Selector, kogitoApp)
	d.addLabelsFromDockerImage(dc, dockerImage)
	d.discoverPortsAndProbesFromImage(dc, dockerImage)
	d.setReplicas(kogitoApp, dc)

	return dc, nil
}

// checkDeploymentDependencies sanity check to create the DeploymentConfig properly
func (d *deploymentConfigResource) checkDeploymentDependencies(bc *buildv1.BuildConfig, sa *corev1.ServiceAccount) (err error) {
	if bc == nil {
		return fmt.Errorf("Impossible to create the DeploymentConfig without a reference to a the service BuildConfig")
	} else if sa == nil {
		return fmt.Errorf("Impossible to create the DeploymentConfig without a reference to a the Kogito ServiceAccount")
	}

	return nil
}

// setReplicas defines the number of container replicas that this DeploymentConfig will have
func (d *deploymentConfigResource) setReplicas(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	replicas := defaultReplicas
	if kogitoApp.Spec.Replicas != nil {
		replicas = *kogitoApp.Spec.Replicas
	}
	dc.Spec.Replicas = replicas
}

// addLabelsFromDockerImage retrieves org.kie labels from DockerImage and adds them to the DeploymentConfig
func (d *deploymentConfigResource) addLabelsFromDockerImage(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) {
	if !d.dockerImageHasLabels(dockerImage) {
		return
	}
	for key, value := range dockerImage.Config.Labels {
		if strings.Contains(key, orgKieNamespaceLabelKey) {
			splitedKey := strings.Split(key, labelNamespaceSep)
			importedKey := splitedKey[len(splitedKey)-1]
			dc.Labels[importedKey] = value
			dc.Spec.Selector[importedKey] = value
			dc.Spec.Template.Labels[importedKey] = value
		}
	}
}

// discoverPortsAndProbesFromImage set Ports and Probes based on labels set on the DockerImage of this DeploymentConfig
func (d *deploymentConfigResource) discoverPortsAndProbesFromImage(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) {
	if !d.dockerImageHasLabels(dockerImage) {
		return
	}
	containerPorts := []corev1.ContainerPort{}
	var nonSecureProbe *corev1.Probe
	for key, value := range dockerImage.Config.Labels {
		if key == labelExposeServices {
			services := strings.Split(value, dockerLabelServicesSep)
			for _, service := range services {
				ports := strings.Split(service, portSep)
				if len(ports) == 0 {
					log.Warnf(portFormatWrongMessage, service)
					continue
				}
				portNumber, err := strconv.Atoi(strings.Split(service, portSep)[0])
				if err != nil {
					log.Warnf(portFormatWrongMessage, service)
					continue
				}
				portName := ports[1]
				containerPorts = append(containerPorts, corev1.ContainerPort{Name: portName, ContainerPort: int32(portNumber), Protocol: corev1.ProtocolTCP})
				// we have at least one service exported using default HTTP protocols, let's used as a probe!
				if portName == defaultExportedProtocol {
					nonSecureProbe = defaultProbe
					nonSecureProbe.Handler.TCPSocket = &corev1.TCPSocketAction{Port: intstr.FromInt(portNumber)}
				}
			}
			break
		}
	}
	// set the ports we've found
	if len(containerPorts) != 0 {
		dc.Spec.Template.Spec.Containers[0].Ports = containerPorts
		if nonSecureProbe != nil {
			dc.Spec.Template.Spec.Containers[0].LivenessProbe = nonSecureProbe
			dc.Spec.Template.Spec.Containers[0].ReadinessProbe = nonSecureProbe
		}
	}
}

func (d *deploymentConfigResource) dockerImageHasLabels(dockerImage *dockerv10.DockerImage) bool {
	if dockerImage == nil || dockerImage.Config == nil || dockerImage.Config.Labels == nil {
		log.Warn("Not found any labels on DockerImage")
		return false
	}
	return true
}
