package definitions

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
)

const (
	kind               = "BuildConfig"
	kindImageStreamTag = "ImageStreamTag"
	tagLatest          = "latest"
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "0.2.0"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
	// S2IBuildType source to image build type will take a source code and transform it into an executable service
	S2IBuildType BuildType = "s2i"
	// RunnerBuildType will create a image with a Kogito Service available
	RunnerBuildType BuildType = "runner"
)

// BuildImageStreams are the image streams needed to perform the initial builds
var BuildImageStreams = map[BuildType]map[v1alpha1.RuntimeType]v1alpha1.Image{
	S2IBuildType: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	RunnerBuildType: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
}

// BuildType which build can we perform? Supported are s2i and runner
type BuildType string

// BuildConfigComposition is the composition of the build configuration for the Kogito App
type BuildConfigComposition struct {
	BuildS2I    buildv1.BuildConfig
	BuildRunner buildv1.BuildConfig
	AsMap       map[BuildType]*buildv1.BuildConfig
}

type buildConfigContext struct {
	Image v1alpha1.Image
}

type buildConfigResource struct {
}

// New creates a new composite build configuration for Kogito App: s2i and runner builds
func (b *buildConfigResource) New(kogitoApp *v1alpha1.KogitoApp) (buildConfig BuildConfigComposition, err error) {
	buildConfig = BuildConfigComposition{}

	buildConfigS2I := buildConfigS2IResource{Image: BuildImageStreams[S2IBuildType][kogitoApp.Spec.Runtime]}
	buildConfigRunner := buildConfigRunnerResource{Image: BuildImageStreams[RunnerBuildType][kogitoApp.Spec.Runtime]}

	if buildConfig.BuildS2I, err = buildConfigS2I.New(kogitoApp); err != nil {
		return buildConfig, err
	}
	if buildConfig.BuildRunner, err = buildConfigRunner.New(kogitoApp, &buildConfig.BuildS2I); err != nil {
		return buildConfig, err
	}

	// transform the builds to a map to facilitate the redesign on controller side.
	// we should remove it after having inventory package to handle the objects
	buildConfig.AsMap = map[BuildType]*buildv1.BuildConfig{
		S2IBuildType:    &buildConfig.BuildS2I,
		RunnerBuildType: &buildConfig.BuildRunner,
	}
	return buildConfig, err
}
