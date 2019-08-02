package definitions

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kogitoApp = &v1alpha1.KogitoApp{
	Spec: v1alpha1.KogitoAppSpec{
		Name: "test",
	},
}

func Test_addDefaultMeta_whenLabelsAreNotDefined(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	addDefaultMeta(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[labelAppName] == "test")
}

func Test_addDefaultMeta_whenAlreadyHasAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Annotations: map[string]string{
			"test": "test",
		},
	}
	addDefaultMeta(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Annotations["test"] == "test")
	assert.True(t, objectMeta.Annotations["org.kie.kogito/managed-by"] == "Kogito Operator")
}
