package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/k8s/clientset/fake"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

var wf = &v1alpha1.Workflow{
	ObjectMeta: metav1.ObjectMeta{
		Name: "wf",
	},
	Spec: v1alpha1.WorkflowSpec{
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("64Mi"),
				corev1.ResourceCPU:    resource.MustParse("50m"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("256Mi"),
				corev1.ResourceCPU:    resource.MustParse("200m"),
			},
		},
		Stages: []v1alpha1.StageItem{
			{
				Name: "stage1",
			},
			{
				Name: "stage2",
				Artifacts: []v1alpha1.ArtifactItem{
					{
						Name:   "art1",
						Source: "stage1/art1",
					},
				},
			},
		},
	},
}

var wfr = &v1alpha1.WorkflowRun{
	ObjectMeta: metav1.ObjectMeta{
		Name: "wfr",
	},
	Spec: v1alpha1.WorkflowRunSpec{
		Stages: []v1alpha1.ParameterConfig{
			{
				Name: "stage1",
				Parameters: []v1alpha1.ParameterItem{
					{
						Name:  "image",
						Value: "busybox:latest",
					},
					{
						Name:  "p1",
						Value: "v1",
					},
				},
			},
		},
		PresetVolumes: []v1alpha1.PresetVolume{
			{
				Type:      v1alpha1.PresetVolumeTypePV,
				Path:      "etc",
				MountPath: "/tmp",
			},
			{
				Type:      v1alpha1.PresetVolumeTypeHostPath,
				Path:      "/etc",
				MountPath: "/tmp",
			},
		},
	},
}

type PodBuilderSuite struct {
	suite.Suite
	client clientset.Interface
}

func (suite *PodBuilderSuite) SetupTest() {
	client := fake.NewSimpleClientset()
	client.PrependReactor("get", "resources", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		if getAction, ok := action.(k8stesting.GetActionImpl); ok {
			name := getAction.GetName()
			switch name {
			case "git":
				return true, &v1alpha1.Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.ResourceSpec{
						Type: v1alpha1.GitResourceType,
						Parameters: []v1alpha1.ParameterItem{
							{
								Name:  "p1",
								Value: "v1",
							},
						},
					},
				}, nil
			case "git-persistent":
				return true, &v1alpha1.Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.ResourceSpec{
						Type: v1alpha1.GitResourceType,
						Persistent: &v1alpha1.Persistent{
							PVC:  "persistent-pvc",
							Path: "/persistent",
						},
					},
				}, nil
			case "image":
				return true, &v1alpha1.Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.ResourceSpec{
						Type: v1alpha1.ImageResourceType,
						Parameters: []v1alpha1.ParameterItem{
							{
								Name:  "IMAGE",
								Value: "busybox:latest",
							},
						},
					},
				}, nil
			}
		}
		return true, nil, nil
	})
	client.PrependReactor("get", "stages", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		if getAction, ok := action.(k8stesting.GetActionImpl); ok {
			name := getAction.GetName()
			switch name {
			case "no-pod":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{},
				}, nil
			case "multi-workload":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{
						Pod: &v1alpha1.PodWorkload{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "c1",
									},
									{
										Name: "c2",
									},
								},
							},
						},
					},
				}, nil
			case "simple":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{
						Pod: &v1alpha1.PodWorkload{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "c1",
									},
								},
							},
						},
					},
				}, nil
			case "unresolvable-argument":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{
						Pod: &v1alpha1.PodWorkload{
							Inputs: v1alpha1.Inputs{
								Arguments: []v1alpha1.ArgumentValue{
									{
										Name: "undefined-arg",
									},
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "c1",
										Image: "{{ undefined-arg }}",
									},
								},
							},
						},
					},
				}, nil
			case "stage1":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{
						Pod: &v1alpha1.PodWorkload{
							Inputs: v1alpha1.Inputs{
								Arguments: []v1alpha1.ArgumentValue{
									{
										Name:  "image",
										Value: "alpine:latest",
									},
									{
										Name:  "dir",
										Value: "/default",
									},
									{
										Name:  "p1",
										Value: "default1",
									},
								},
								Resources: []v1alpha1.ResourceItem{
									{
										Name: "git",
										Path: "/resource",
									},
								},
							},
							Outputs: v1alpha1.Outputs{
								Artifacts: []v1alpha1.ArtifactItem{
									{
										Name: "art1",
										Path: "/tmp/artifact.tar",
									},
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:       "c1",
										Image:      "{{ image }}",
										WorkingDir: "{{ dir }}",
									},
									{
										Name: "wsc-c2",
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceCPU: resource.MustParse("100m"),
											},
											Limits: corev1.ResourceList{
												corev1.ResourceMemory: resource.MustParse("256Mi"),
											},
										},
									},
								},
							},
						},
					},
				}, nil
			case "stage2":
				return true, &v1alpha1.Stage{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
					Spec: v1alpha1.StageSpec{
						Pod: &v1alpha1.PodWorkload{
							Inputs: v1alpha1.Inputs{
								Resources: []v1alpha1.ResourceItem{
									{
										Name: "git-persistent",
										Path: "/resource",
									},
								},
								Artifacts: []v1alpha1.ArtifactItem{
									{
										Name: "art1",
										Path: "/tmp/art1",
									},
								},
							},
							Outputs: v1alpha1.Outputs{
								Resources: []v1alpha1.ResourceItem{
									{
										Name: "image",
										Path: "/workspace/image.tar",
									},
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "c1",
										Image: "busybox:v1.0",
									},
								},
							},
						},
					},
				}, nil
			default:
				return true, nil, errors.NewNotFound(action.GetResource().GroupResource(), name)
			}
		}
		return true, nil, nil
	})
	suite.client = client
}

func getStage(client clientset.Interface, stage string) *v1alpha1.Stage {
	stg, _ := client.CycloneV1alpha1().Stages("").Get(stage, metav1.GetOptions{})
	return stg
}

func (suite *PodBuilderSuite) TestPrepare() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "simple"))
	err := builder.Prepare()
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), builder.pod.Name)
	assert.NotEmpty(suite.T(), builder.pod.Labels[meta.LabelWorkflowRunName])
	assert.Equal(suite.T(), "simple", builder.pod.Annotations[meta.AnnotationStageName])
	assert.Equal(suite.T(), "wfr", builder.pod.Annotations[meta.AnnotationWorkflowRunName])
}

func (suite *PodBuilderSuite) TestResolveArguments() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "unresolvable-argument"))
	err := builder.Prepare()
	assert.Nil(suite.T(), err)
	err = builder.ResolveArguments()
	assert.Error(suite.T(), err)

	builder = NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	err = builder.Prepare()
	assert.Nil(suite.T(), err)
	err = builder.ResolveArguments()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "busybox:latest", builder.pod.Spec.Containers[0].Image)
	assert.Equal(suite.T(), "/default", builder.pod.Spec.Containers[0].WorkingDir)
	assert.Equal(suite.T(), corev1.RestartPolicyNever, builder.pod.Spec.RestartPolicy)
}

func (suite *PodBuilderSuite) TestCreateVolumes() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	err := builder.Prepare()
	assert.Nil(suite.T(), err)
	err = builder.ResolveArguments()
	assert.Nil(suite.T(), err)
	err = builder.CreateVolumes()
	assert.Nil(suite.T(), err)

	var volumes []string
	for _, v := range builder.pod.Spec.Volumes {
		volumes = append(volumes, v.Name)
	}

	assert.Contains(suite.T(), volumes, common.CoordinatorSidecarVolumeName)
	assert.Contains(suite.T(), volumes, common.PresetVolumeName(1))
	assert.NotContains(suite.T(), volumes, common.DockerInDockerSockVolume)
	assert.NotContains(suite.T(), volumes, common.DefaultPvVolumeName)
	assert.NotContains(suite.T(), volumes, common.DockerConfigJSONVolume)
}

func (suite *PodBuilderSuite) TestCreatePVCVolume() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	assert.Equal(suite.T(), "v1", builder.CreatePVCVolume("v1", "pvc1"))
	assert.Equal(suite.T(), "v1", builder.CreatePVCVolume("v2", "pvc1"))
}

func (suite *PodBuilderSuite) TestResolveInputResources() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveInputResources())
	initContainer := builder.pod.Spec.InitContainers[0]
	assert.Equal(suite.T(), "i1", initContainer.Name)
	assert.Equal(suite.T(), GetResourceVolumeName("git"), initContainer.VolumeMounts[0].Name)
	assert.Equal(suite.T(), "", initContainer.VolumeMounts[0].SubPath)
	var envs []string
	for _, e := range initContainer.Env {
		envs = append(envs, e.Name)
	}
	assert.Contains(suite.T(), envs, "p1")

	builder = NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage2"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Nil(suite.T(), builder.ResolveInputResources())
	initContainer = builder.pod.Spec.InitContainers[0]
	assert.Equal(suite.T(), "i1", initContainer.Name)
	assert.Equal(suite.T(), common.InputResourceVolumeName("git-persistent"), initContainer.VolumeMounts[0].Name)
	assert.Equal(suite.T(), "/persistent", initContainer.VolumeMounts[0].SubPath)
	var resourceMount corev1.VolumeMount
	for _, vm := range builder.pod.Spec.Containers[0].VolumeMounts {
		if vm.Name == common.InputResourceVolumeName("git-persistent") {
			resourceMount = vm
		}
	}
	assert.Equal(suite.T(), corev1.VolumeMount{
		Name:      common.InputResourceVolumeName("git-persistent"),
		MountPath: "/resource",
		SubPath:   "/persistent/data",
	}, resourceMount)
}

func (suite *PodBuilderSuite) TestResolveOutputResources() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage2"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Nil(suite.T(), builder.ResolveOutputResources())

	var sidecar corev1.Container
	for _, c := range builder.pod.Spec.Containers {
		if c.Name == OutputContainerName(1) {
			sidecar = c
			break
		}
	}
	assert.Equal(suite.T(), []string{"push"}, sidecar.Args)
	var envs []string
	for _, e := range sidecar.Env {
		envs = append(envs, e.Name)
	}
	assert.Contains(suite.T(), envs, "IMAGE")
	assert.Equal(suite.T(), corev1.PullIfNotPresent, sidecar.ImagePullPolicy)
	var vms []string
	var mountPaths []string
	for _, m := range sidecar.VolumeMounts {
		vms = append(vms, m.Name)
		mountPaths = append(mountPaths, m.MountPath)
	}
	assert.Contains(suite.T(), vms, common.CoordinatorSidecarVolumeName)
	assert.Contains(suite.T(), mountPaths, common.ResolverDefaultDataPath)
	assert.Contains(suite.T(), vms, common.DockerInDockerSockVolume)
	assert.Contains(suite.T(), mountPaths, common.ResolverNotifyDirPath)
}

func (suite *PodBuilderSuite) TestResolveInputArtifacts() {
	controller.Config = controller.WorkflowControllerConfig{}
	defer func() {
		controller.Config = controller.WorkflowControllerConfig{}
	}()

	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage2"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Error(suite.T(), builder.ResolveInputArtifacts())

	builder.executionContext.PVC = "pvc1"
	assert.Nil(suite.T(), builder.ResolveInputArtifacts())

	for _, c := range builder.pod.Spec.Containers {
		assert.Contains(suite.T(), c.VolumeMounts, corev1.VolumeMount{
			Name:      common.DefaultPvVolumeName,
			MountPath: "/tmp/art1",
			SubPath:   common.ArtifactPath("wfr", "stage1", "art1") + "/artifact.tar",
		})
	}
}

func (suite *PodBuilderSuite) TestApplyResourceRequirements() {
	configured := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("32Mi"),
			corev1.ResourceCPU:    resource.MustParse("25m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("32Mi"),
			corev1.ResourceCPU:    resource.MustParse("25m"),
		},
	}
	controller.Config = controller.WorkflowControllerConfig{}
	controller.Config.ResourceRequirements = configured
	defer func() {
		controller.Config = controller.WorkflowControllerConfig{}
	}()

	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Nil(suite.T(), builder.ResolveInputResources())

	assert.Nil(suite.T(), builder.ApplyResourceRequirements())
	for _, c := range builder.pod.Spec.InitContainers {
		assert.Equal(suite.T(), configured, c.Resources)
	}

	for _, c := range builder.pod.Spec.Containers {
		if c.Name == "c1" {
			assert.Equal(suite.T(), corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("64Mi"),
					corev1.ResourceCPU:    resource.MustParse("50m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("256Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			}, c.Resources)
		}

		if c.Name == "workload-sidecar-c2" {
			assert.Equal(suite.T(), corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("64Mi"),
					corev1.ResourceCPU:    resource.MustParse("100m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("256Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			}, c.Resources)
		}
	}
}

func (suite *PodBuilderSuite) TestArtifactFileName() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage2"))
	name, _ := builder.ArtifactFileName("stage1", "art1")
	assert.Equal(suite.T(), "artifact.tar", name)
}

func (suite *PodBuilderSuite) TestAddCommonVolumes() {
	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage2"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Nil(suite.T(), builder.CreateVolumes())
	assert.Nil(suite.T(), builder.MountPresetVolumes())

	for _, c := range builder.pod.Spec.Containers {
		assert.Contains(suite.T(), c.VolumeMounts, corev1.VolumeMount{
			Name:      common.DefaultPvVolumeName,
			SubPath:   "etc",
			MountPath: "/tmp",
		})

		assert.Contains(suite.T(), c.VolumeMounts, corev1.VolumeMount{
			Name:      common.PresetVolumeName(1),
			MountPath: "/tmp",
			ReadOnly:  true,
		})
	}
}

func TestPodBuilderSuite(t *testing.T) {
	suite.Run(t, new(PodBuilderSuite))
}
