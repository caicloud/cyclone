package pod

import (
	"encoding/json"
	"fmt"
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

var tmp1 = "busybox:latest"
var tmp2 = "v1"
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
						Value: &tmp1,
					},
					{
						Name:  "p1",
						Value: &tmp2,
					},
				},
			},
		},
		PresetVolumes: []v1alpha1.PresetVolume{
			{
				Type:      v1alpha1.PresetVolumeTypePVC,
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
	v1 := "v1"
	v2 := "busybox:latest"
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
								Value: &v1,
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
								Value: &v2,
							},
						},
					},
				}, nil
			}
		}
		return true, nil, nil
	})

	client.PrependReactor("list", "resources", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &v1alpha1.ResourceList{
			Items: []v1alpha1.Resource{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "resource-type-git",
						Labels: map[string]string{
							"resource.cyclone.dev/template": "true",
						},
					},
					Spec: v1alpha1.ResourceSpec{
						Type:     v1alpha1.GitResourceType,
						Resolver: "cyclone-resolver-git",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "resource-type-image",
						Labels: map[string]string{
							"resource.cyclone.dev/template": "true",
						},
					},
					Spec: v1alpha1.ResourceSpec{
						Type:     v1alpha1.ImageResourceType,
						Resolver: "cyclone-resolver-image",
					},
				},
			},
		}, nil
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
				v1 := "alpine:latest"
				v2 := "/default"
				v3 := "default1"
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
										Value: &v1,
									},
									{
										Name:  "dir",
										Value: &v2,
									},
									{
										Name:  "p1",
										Value: &v3,
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
	controller.Config = controller.WorkflowControllerConfig{
		Images: map[string]string{
			controller.ToolboxImage: "cyclone-toolbox:v0.1",
		},
	}
	defer func() {
		controller.Config = controller.WorkflowControllerConfig{}
	}()

	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveInputResources())
	initContainer := builder.pod.Spec.InitContainers[1]
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
	initContainer = builder.pod.Spec.InitContainers[1]
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
	assert.Nil(suite.T(), builder.InjectEnvs())

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
	assert.Contains(suite.T(), envs, common.EnvMetadataNamespace)
	assert.Contains(suite.T(), envs, common.EnvWorkflowName)
	assert.Contains(suite.T(), envs, common.EnvWorkflowrunName)
	assert.Contains(suite.T(), envs, common.EnvStageName)
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
	controller.Config = controller.WorkflowControllerConfig{
		Images: map[string]string{
			controller.ToolboxImage: "cyclone-toolbox:v0.1",
		},
	}
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

func checkQuantity(t *testing.T, desc string, expected, got corev1.ResourceList) {
	t.Helper()
	for name, expectedVal := range expected {
		gotVal := got[name]
		if !expectedVal.Equal(gotVal) {
			t.Errorf("%s: expected=%s, got=%s", desc, expectedVal.String(), gotVal.String())
		}
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
	controller.Config = controller.WorkflowControllerConfig{
		Images: map[string]string{
			controller.ToolboxImage: "cyclone-toolbox:v0.1",
		},
	}
	controller.Config.ResourceRequirements = configured
	defer func() {
		controller.Config = controller.WorkflowControllerConfig{}
	}()

	builder := NewBuilder(suite.client, wf, wfr, getStage(suite.client, "stage1"))
	assert.Nil(suite.T(), builder.Prepare())
	assert.Nil(suite.T(), builder.ResolveArguments())
	assert.Nil(suite.T(), builder.ResolveInputResources())

	// controller.Config.CustomContainerCPUWeight defaults to 100
	assert.Nil(suite.T(), builder.ApplyResourceRequirements())
	for _, c := range builder.pod.Spec.InitContainers {
		checkQuantity(suite.T(), fmt.Sprintf("%s: requests", c.Name), wf.Spec.Resources.Requests, c.Resources.Requests)
		checkQuantity(suite.T(), fmt.Sprintf("%s: limits", c.Name), wf.Spec.Resources.Limits, c.Resources.Limits)
	}

	expectedContainerResources := map[string]corev1.ResourceRequirements{
		"c1": {
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("32Mi"),
				corev1.ResourceCPU:    resource.MustParse("25m"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("128Mi"),
				corev1.ResourceCPU:    resource.MustParse("100m"),
			},
		},
		"wsc-c2": {
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("32Mi"),
				corev1.ResourceCPU:    resource.MustParse("100m"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("256Mi"),
				corev1.ResourceCPU:    resource.MustParse("100m"),
			},
		},
	}
	for _, c := range builder.pod.Spec.Containers {
		expected := expectedContainerResources[c.Name]
		checkQuantity(suite.T(), fmt.Sprintf("%s: requests", c.Name), expected.Requests, c.Resources.Requests)
		checkQuantity(suite.T(), fmt.Sprintf("%s: limits", c.Name), expected.Limits, c.Resources.Limits)
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

func Test_applyResourceRequirements_cpu_format(t *testing.T) {
	containers := []corev1.Container{
		{Name: "w1"},
		{Name: "csc-c1"},
	}
	requirements := &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
			corev1.ResourceCPU:    resource.MustParse("100m"),
		},
	}
	updated := applyResourceRequirements(containers, requirements, true, 50)
	got, err := json.Marshal(updated)
	const expected = `[{"name":"w1","resources":{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"50m","memory":"128Mi"}}},{"name":"csc-c1","resources":{"limits":{"cpu":"100m","memory":"0"},"requests":{"cpu":"50m","memory":"0"}}}]`
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, string(got))
}

func TestApplyResourceRequirements(t *testing.T) {
	testCases := map[string]struct {
		containers          []corev1.Container
		requirements        *corev1.ResourceRequirements
		averageToContainers bool
		customCPUWeight     int
		expects             []corev1.Container
	}{
		"t1": {
			containers: []corev1.Container{
				{
					Name: "c1",
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
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("128Mi"),
					corev1.ResourceCPU:    resource.MustParse("100m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("256Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     100,
			expects: []corev1.Container{
				{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("64Mi"),
							corev1.ResourceCPU:    resource.MustParse("50m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("128Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
					},
				},
				{
					Name: "wsc-c2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("64Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("256Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
					},
				},
			},
		},
		"t2": {
			containers: []corev1.Container{
				{
					Name: "c1",
				},
				{
					Name:      "csc-c2",
					Resources: corev1.ResourceRequirements{},
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("128Mi"),
					corev1.ResourceCPU:    resource.MustParse("100m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("256Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     100,
			expects: []corev1.Container{
				{
					Name: "c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("128Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("256Mi"),
							corev1.ResourceCPU:    resource.MustParse("200m"),
						},
					},
				},
				{
					Name: "csc-c2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("0"),
							corev1.ResourceCPU:    resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("0"),
							corev1.ResourceCPU:    resource.MustParse("200m"),
						},
					},
				},
			},
		},
		"t3": {
			containers: []corev1.Container{
				{
					Name:      "csc-c2",
					Resources: corev1.ResourceRequirements{},
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("128Mi"),
					corev1.ResourceCPU:    resource.MustParse("100m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("256Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     100,
			expects: []corev1.Container{
				{
					Name: "csc-c2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("0"),
							corev1.ResourceCPU:    resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("0"),
							corev1.ResourceCPU:    resource.MustParse("200m"),
						},
					},
				},
			},
		},
		"multiple-custom-containers-no-avg-weight-100(initContainers)": {
			containers: []corev1.Container{
				{
					Name: "w1",
				},
				{
					Name: "w2",
				},
				{
					Name: "w3",
				},
				{
					Name: "csc-c1",
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			averageToContainers: false,
			customCPUWeight:     100,
			expects: []corev1.Container{
				{
					Name: "w1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "w2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "w3",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "csc-c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("0"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
			},
		},
		"single-custom-container-avg-weight-60": {
			containers: []corev1.Container{
				{
					Name: "w1",
				},
				{
					Name: "csc-dind",
				},
				{
					Name: "csc-c1",
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     60,
			expects: []corev1.Container{
				{
					Name: "w1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("600m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "csc-dind",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
				{
					Name: "csc-c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
			},
		},
		"multiple-custom-containers-avg-weight-50": {
			containers: []corev1.Container{
				{
					Name: "w1",
				},
				{
					Name: "w2",
				},
				{
					Name: "csc-c1",
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     50,
			expects: []corev1.Container{
				{
					Name: "w1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
				{
					Name: "w2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
				{
					Name: "csc-c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
			},
		},
		"resource-not-divisible": {
			containers: []corev1.Container{
				{
					Name: "w1",
				},
				{
					Name: "csc-c1",
				},
				{
					Name: "csc-c2",
				},
				{
					Name: "csc-c3",
				},
			},
			requirements: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
			averageToContainers: true,
			customCPUWeight:     50,
			expects: []corev1.Container{
				{
					Name: "w1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1000m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				{
					Name: "csc-c1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("166m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
				{
					Name: "csc-c2",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("166m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
				{
					Name: "csc-c3",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("166m"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("0"),
						},
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			results := applyResourceRequirements(testCase.containers, testCase.requirements, testCase.averageToContainers, testCase.customCPUWeight)
			for _, c := range results {
				for _, ec := range testCase.expects {
					if c.Name == ec.Name {
						checkQuantity(t, fmt.Sprintf("%s: requests", c.Name), ec.Resources.Requests, c.Resources.Requests)
						checkQuantity(t, fmt.Sprintf("%s: limits", c.Name), ec.Resources.Limits, c.Resources.Limits)
					}
				}
			}
		})
	}
}
