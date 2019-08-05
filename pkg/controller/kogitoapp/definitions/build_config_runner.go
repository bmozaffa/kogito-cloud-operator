package definitions

import (
	"errors"
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	destinationDir   = "."
	runnerSourcePath = "/home/kogito/bin"
)

type buildConfigRunnerResource struct {
	Image v1alpha1.Image
}

func (b *buildConfigRunnerResource) New(kogitoApp *v1alpha1.KogitoApp, fromBuild *buildv1.BuildConfig) (buildConfig buildv1.BuildConfig, err error) {
	if fromBuild == nil {
		err = errors.New("Impossible to create a runner build configuration without a s2i build definition")
		return buildConfig, err
	}
	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Spec.Name,
			Namespace: kogitoApp.Namespace,
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", kogitoApp.Spec.Name, tagLatest)}
	b.setBuildSource(kogitoApp, &buildConfig, fromBuild)
	b.setBuildStrategy(kogitoApp, &buildConfig)
	b.seTriggers(&buildConfig, fromBuild)
	setGroupVersionKind(&buildConfig.TypeMeta, BuildConfigKind)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig, err
}

func (b *buildConfigRunnerResource) setBuildSource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.Type = buildv1.BuildSourceImage
	buildConfig.Spec.Source.Images = []buildv1.ImageSource{
		{
			From: *fromBuildConfig.Spec.Output.To,
			Paths: []buildv1.ImageSourcePath{
				{
					DestinationDir: destinationDir,
					SourcePath:     runnerSourcePath,
				},
			},
		},
	}
}

func (b *buildConfigRunnerResource) setBuildStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", b.Image.ImageStreamName, b.Image.ImageStreamTag),
			Namespace: b.Image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
	}
}

func (b *buildConfigRunnerResource) seTriggers(buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
		{
			Type:        buildv1.ImageChangeBuildTriggerType,
			ImageChange: &buildv1.ImageChangeTrigger{From: fromBuildConfig.Spec.Output.To},
		},
	}
}
