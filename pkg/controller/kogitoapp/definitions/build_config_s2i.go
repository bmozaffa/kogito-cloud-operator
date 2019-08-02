package definitions

import (
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nameSuffix = "-builder"
)

type buildConfigS2IResource struct {
	Image v1alpha1.Image
}

func (b *buildConfigS2IResource) New(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig) {
	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", kogitoApp.Spec.Name, nameSuffix),
			Namespace: kogitoApp.Namespace,
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}
	buildConfig.SetGroupVersionKind(buildv1.SchemeGroupVersion.WithKind(kind))
	b.setBuildSource(kogitoApp, &buildConfig)
	b.setBuildStrategy(kogitoApp, &buildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig
}

func (b *buildConfigS2IResource) setBuildSource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.ContextDir = kogitoApp.Spec.Build.GitSource.ContextDir
	buildConfig.Spec.Source.Git = &buildv1.GitBuildSource{
		URI: *kogitoApp.Spec.Build.GitSource.URI,
		Ref: kogitoApp.Spec.Build.GitSource.Reference,
	}
}

func (b *buildConfigS2IResource) setBuildStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", b.Image.ImageStreamName, b.Image.ImageStreamTag),
			Namespace: b.Image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
		Env:         shared.FromEnvToEnvVar(kogitoApp.Spec.Build.Env),
		Incremental: &kogitoApp.Spec.Build.Incremental,
	}
}
