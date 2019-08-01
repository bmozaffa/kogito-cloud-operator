package definitions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_addDefaultMeta_whenLabelsAreNotDefined(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	addDefaultMeta(objectMeta, "test")
	assert.True(t, objectMeta.Labels[labelAppName] == "test")
}

func Test_addDefaultMeta_whenAlreadyHasAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Annotations: map[string]string{
			"test": "test",
		},
	}
	addDefaultMeta(objectMeta, "test")
	assert.True(t, objectMeta.Annotations["test"] == "test")
	assert.True(t, objectMeta.Annotations["org.kie.kogito/managed-by"] == "Kogito Operator")
}
