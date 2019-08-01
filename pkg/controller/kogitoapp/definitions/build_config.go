package definitions

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	buildS2INameSuffix        = "-builder"
	buildRunnerDestinationDir = "."
	buildRunnerSourcePath     = "/home/kogito/bin"
	kind                      = "BuildConfig"
	kindImageStreamTag        = "ImageStreamTag"
	tagLatest                 = "latest"
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
}

type innerBuildConfig struct {
	Suffix string
	Image  v1alpha1.Image
	Type   BuildType
}

type buildConfigResource struct {
}

func (b *buildConfigResource) New(kogitoApp *v1alpha1.KogitoApp) (buildConfig BuildConfigComposition, err error) {
	// setup both builders
	buildConfigS2I := innerBuildConfig{
		Suffix: buildS2INameSuffix,
		Image:  BuildImageStreams[S2IBuildType][kogitoApp.Spec.Runtime],
		Type:   S2IBuildType,
	}
	buildConfigRunner := innerBuildConfig{
		Suffix: "",
		Image:  BuildImageStreams[RunnerBuildType][kogitoApp.Spec.Runtime],
		Type:   RunnerBuildType,
	}
	buildConfig = BuildConfigComposition{}
	buildConfig.BuildS2I, err = buildConfigS2I.new(kogitoApp, nil)
	buildConfig.BuildRunner, err = buildConfigRunner.new(kogitoApp, &buildConfig.BuildS2I)
	return buildConfig, err
}

func (b *innerBuildConfig) new(kogitoApp *v1alpha1.KogitoApp, fromBuildConfig *buildv1.BuildConfig) (buildConfig buildv1.BuildConfig, err error) {
	if b.Type == RunnerBuildType && fromBuildConfig == nil {
		err = errors.New("Impossible to create a runner build configuration without a s2i build definition")
		return buildConfig, err
	}
	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", kogitoApp.Spec.Name, b.Suffix),
			Namespace: kogitoApp.Namespace,
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}
	buildConfig.SetGroupVersionKind(buildv1.SchemeGroupVersion.WithKind(kind))
	// detailed configuration
	b.setBuildSource(kogitoApp, &buildConfig, fromBuildConfig)
	b.setBuildStrategy(kogitoApp, &buildConfig)
	b.seTriggers(&buildConfig, fromBuildConfig)
	// meta
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp.Name)
	// end
	return buildConfig, nil
}

func (b *innerBuildConfig) setBuildSource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	imageOutput := ""
	if b.Type == S2IBuildType {
		// From Git build
		imageOutput = buildConfig.Name
		buildConfig.Spec.Source.ContextDir = kogitoApp.Spec.Build.GitSource.ContextDir
		buildConfig.Spec.Source.Git = &buildv1.GitBuildSource{
			URI: *kogitoApp.Spec.Build.GitSource.URI,
			Ref: kogitoApp.Spec.Build.GitSource.Reference,
		}
	} else {
		// From s2i build
		imageOutput = kogitoApp.Spec.Name
		buildConfig.Spec.Source.Type = buildv1.BuildSourceImage
		buildConfig.Spec.Source.Images = []buildv1.ImageSource{
			{
				From: *fromBuildConfig.Spec.Output.To,
				Paths: []buildv1.ImageSourcePath{
					{
						DestinationDir: buildRunnerDestinationDir,
						SourcePath:     buildRunnerSourcePath,
					},
				},
			},
		}
	}

	buildConfig.Spec.Output.To = &corev1.ObjectReference{Name: fmt.Sprintf("%s:%s", imageOutput, tagLatest), Kind: kindImageStreamTag}
}

func (b *innerBuildConfig) setBuildStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", b.Image.ImageStreamName, b.Image.ImageStreamTag),
			Namespace: b.Image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
	}

	if b.Type == S2IBuildType {
		buildConfig.Spec.Strategy.SourceStrategy.Env = shared.FromEnvToEnvVar(kogitoApp.Spec.Build.Env)
		buildConfig.Spec.Strategy.SourceStrategy.Incremental = &kogitoApp.Spec.Build.Incremental
	}
}

func (b *innerBuildConfig) seTriggers(buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	if b.Type == RunnerBuildType {
		buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
			{
				Type:        buildv1.ImageChangeBuildTriggerType,
				ImageChange: &buildv1.ImageChangeTrigger{From: fromBuildConfig.Spec.Output.To},
			},
		}
	}
}
