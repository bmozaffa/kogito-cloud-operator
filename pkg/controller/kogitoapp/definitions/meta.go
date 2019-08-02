package definitions

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var commonAnnotations = map[string]string{
	"org.kie.kogito/managed-by":   "Kogito Operator",
	"org.kie.kogito/operator-crd": "KogitoApp",
}

const (
	labelAppName = "app"
)

func addDefaultMeta(objectMeta *metav1.ObjectMeta, kogitoApp *v1alpha1.KogitoApp) {
	if objectMeta != nil {
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = map[string]string{}
		}
		if objectMeta.Labels == nil {
			objectMeta.Labels = map[string]string{}
		}
		for key, value := range commonAnnotations {
			objectMeta.Annotations[key] = value
		}
		objectMeta.Labels[labelAppName] = kogitoApp.Spec.Name
	}
}
