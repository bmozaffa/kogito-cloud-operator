package definitions

import (
	"testing"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	dockerv10 "github.com/openshift/api/image/docker10"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_serviceResource_NewWithAndWithoutDockerImg(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Name: "test",
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
			},
		},
	}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				labelExposeServices:                  "8080:http",
				orgKieNamespaceLabelKey + "operator": "kogito",
			},
		},
	}
	svcResource := &serviceResource{}
	dcResource := &deploymentConfigResource{}
	saResource := &serviceAccountResource{}
	bcResource := &buildConfigResource{}
	bc, _ := bcResource.New(kogitoApp)
	sa := saResource.New(kogitoApp)
	dc, _ := dcResource.New(kogitoApp, &bc.BuildRunner, &sa, nil)
	svc, err := svcResource.New(kogitoApp, dc)
	assert.NotNil(t, err)
	assert.Nil(t, svc)
	// try again, now with ports
	dc, _ = dcResource.New(kogitoApp, &bc.BuildRunner, &sa, dockerImage)
	svc, err = svcResource.New(kogitoApp, dc)
	assert.Nil(t, err)
	assert.NotNil(t, svc)
	assert.Len(t, svc.Spec.Ports, 1)
	assert.Equal(t, int32(8080), svc.Spec.Ports[0].Port)
}
