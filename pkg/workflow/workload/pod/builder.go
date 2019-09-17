package pod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cbroglie/mustache"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	ccommon "github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/common/values"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
	"github.com/caicloud/cyclone/pkg/workflow/values/ref"
)

// Builder is builder used to build pod for stage
type Builder struct {
	client           clientset.Interface
	wf               *v1alpha1.Workflow
	wfr              *v1alpha1.WorkflowRun
	stg              *v1alpha1.Stage
	rendered         *v1alpha1.StageSpec
	stage            string
	pod              *corev1.Pod
	pvcVolumes       map[string]string
	executionContext *v1alpha1.ExecutionContext
	outputResources  []*v1alpha1.Resource
	refProcessor     *ref.Processor
}

// NewBuilder creates a new pod builder.
func NewBuilder(client clientset.Interface, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun, stg *v1alpha1.Stage) *Builder {
	return &Builder{
		client:           client,
		wf:               wf,
		wfr:              wfr,
		stage:            stg.Name,
		stg:              stg,
		pod:              &corev1.Pod{},
		pvcVolumes:       make(map[string]string),
		executionContext: GetExecutionContext(wfr),
		refProcessor:     ref.NewProcessor(wfr),
	}
}

var (
	zeroQuantity = resource.MustParse("0")
)

// Prepare ...
func (m *Builder) Prepare() error {
	if m.stg.Spec.Pod == nil {
		return fmt.Errorf("pod must be defined in stage spec, stage: %s", m.stage)
	}

	// Only one workload container supported, others should be sidecar marked by special
	// container name prefix.
	var workloadContainers int
	for _, c := range m.stg.Spec.Pod.Spec.Containers {
		if !strings.HasPrefix(c.Name, common.WorkloadSidecarPrefix) {
			workloadContainers++
		}
	}
	if workloadContainers != 1 {
		return fmt.Errorf("only one workload containers supported, others should be sidecars, stage: %s", m.stage)
	}

	m.pod.ObjectMeta = metav1.ObjectMeta{
		Name:      Name(m.wf.Name, m.stage),
		Namespace: m.executionContext.Namespace,
		Labels: map[string]string{
			meta.LabelWorkflowRunName: m.wfr.Name,
			meta.LabelPodCreatedBy:    meta.CycloneCreator,
		},
		Annotations: map[string]string{
			meta.AnnotationIstioInject:     meta.AnnotationValueFalse,
			meta.AnnotationWorkflowRunName: m.wfr.Name,
			meta.AnnotationStageName:       m.stage,
			meta.AnnotationMetaNamespace:   m.wfr.Namespace,
		},
	}

	// If controller instance name is set, add label to the pod created.
	if instance := os.Getenv(ccommon.ControllerInstanceEnvName); len(instance) != 0 {
		m.pod.ObjectMeta.Labels[meta.LabelControllerInstance] = instance
	}

	return nil
}

// ResolveArguments ...
func (m *Builder) ResolveArguments() error {
	parameters := make(map[string]string)
	for _, s := range m.wfr.Spec.Stages {
		if s.Name == m.stage {
			for _, p := range s.Parameters {
				if p.Value != nil {
					parameters[p.Name] = *p.Value
				}
			}
		}
	}
	for _, a := range m.stg.Spec.Pod.Inputs.Arguments {
		if _, ok := parameters[a.Name]; !ok {
			if a.Value == nil {
				log.WithField("arg", a.Name).
					WithField("stg", m.stg.Name).
					Error("Argument not set and without default value")
				return fmt.Errorf("argument '%s' not set in stage '%s' and without default value", a.Name, m.stg.Name)
			}
			parameters[a.Name] = *a.Value
		}
	}

	for k, v := range parameters {
		v = values.GenerateValue(v)
		resolved, err := m.refProcessor.ResolveRefStringValue(v, m.client)
		if err != nil {
			log.WithField("key", k).WithField("value", v).Error("resolve ref failed:", err)
			return err
		}
		parameters[k] = resolved
	}

	// Escape double quotes and '\n'
	for k, v := range parameters {
		parameters[k] = strings.Replace(v, "\"", "\\\"", -1)
		parameters[k] = strings.Replace(parameters[k], "\n", "\\n", -1)
	}

	// Handle cases when parameter values containers templates
	for k, v := range parameters {
		rendered, err := mustache.Render(v, parameters)
		if err != nil {
			return err
		}
		parameters[k] = rendered
	}

	log.WithField("params", parameters).Debug("Parameters collected")
	b, err := json.Marshal(m.stg.Spec)
	if err != nil {
		return err
	}

	rendered, err := mustache.Render(string(b), parameters)
	if err != nil {
		return err
	}

	renderedSpec := v1alpha1.StageSpec{}
	err = json.Unmarshal([]byte(rendered), &renderedSpec)
	if err != nil {
		log.Errorf("Unmarshal rendered stage spec %s error: %v", rendered, err)
		return err
	}
	m.rendered = &renderedSpec
	m.pod.Spec = renderedSpec.Pod.Spec
	m.pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	return nil
}

// EnsureContainerNames ensures all containers have name set.
func (m *Builder) EnsureContainerNames() error {
	if m.stg.Spec.Pod == nil {
		return nil
	}

	nameMap := make(map[string]struct{})
	for _, c := range m.stg.Spec.Pod.Spec.Containers {
		if len(c.Name) > 0 {
			nameMap[c.Name] = struct{}{}
		}
	}

	nameIndex := 0
	for i := range m.stg.Spec.Pod.Spec.Containers {
		if len(m.stg.Spec.Pod.Spec.Containers[i].Name) > 0 {
			continue
		}

		for {
			if _, ok := nameMap[fmt.Sprintf("c%d", nameIndex)]; !ok {
				break
			}
			nameIndex++
		}
		m.stg.Spec.Pod.Spec.Containers[i].Name = fmt.Sprintf("c%d", nameIndex)
	}

	return nil
}

// CreateVolumes ...
func (m *Builder) CreateVolumes() error {
	// Add emptyDir volume to be shared between coordinator and sidecars, e.g. resource resolvers.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.CoordinatorSidecarVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Add host volume for host docker socket file for coordinator container.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.HostDockerSockVolumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: common.DockerSockFilePath,
			},
		},
	})

	// Add PVC volume to pod if configured.
	if m.executionContext.PVC != "" {
		if n := m.CreatePVCVolume(common.DefaultPvVolumeName, m.executionContext.PVC); n != common.DefaultPvVolumeName {
			log.WithField("volume", n).Error("Another volume already exist for the PVC: ", m.executionContext.PVC)
			return fmt.Errorf("%s already in another volume %s", m.executionContext.PVC, n)
		}
	}

	// Add preset volumes to pod if configured
	for i, v := range m.wfr.Spec.PresetVolumes {
		switch v.Type {
		case v1alpha1.PresetVolumeTypeHostPath:
			m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
				Name: common.PresetVolumeName(i),
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: v.Path,
					},
				},
			})
		case v1alpha1.PresetVolumeTypePVC:
			// Use the default PVC for preset volume, and the volume already created before
		case v1alpha1.PresetVolumeTypeSecret:
			m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
				Name: common.PresetVolumeName(i),
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: *v.ObjectName,
						Items: []corev1.KeyToPath{
							{
								Key:  v.Path,
								Path: v.SubPath,
							},
						},
					},
				},
			})
		case v1alpha1.PresetVolumeTypeConfigMap:
			m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
				Name: common.PresetVolumeName(i),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: *v.ObjectName,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  v.Path,
								Path: v.SubPath,
							},
						},
					},
				},
			})
		default:
			log.WithField("type", v.Type).Warning("Unknown preset volume type.")
		}
	}

	// Create emptyDir volume for cyclone tools. Cyclone will inject some tools to containers via this volume.
	// For example, 'fstream' for git resolver containers to send their logs to log collector (cyclone-server
	// if not configured).
	// Add emptyDir volume to be shared between coordinator and sidecars, e.g. resource resolvers.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.ToolsVolume,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	return nil
}

// CreatePVCVolume tries to create a PVC volume for the given volume name and PVC name.
// If no volume available for the PVC, a new volume would be created and the volume name
// will be returned. If a volume of the given PVC already exists, return name of  the volume,
// note that in this case, the returned volume name is usually different to the provided
// 'volumeName' argument.
func (m *Builder) CreatePVCVolume(volumeName, pvc string) string {
	// PVC --> Volume Name
	if volume, ok := m.pvcVolumes[pvc]; ok {
		return volume
	}

	// Create volume if no volumes available for the PVC.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})

	m.pvcVolumes[pvc] = volumeName
	return volumeName
}

// CreateEmptyDirVolume creates a EmptyDir volume for the pod with the given name
func (m *Builder) CreateEmptyDirVolume(volumeName string) {
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}

// ResolveInputResources creates init containers for each input resource and also mount
// resource to workload containers.
func (m *Builder) ResolveInputResources() error {
	var toolboxImage string
	if img, ok := controller.Config.Images[controller.ToolboxImage]; ok {
		toolboxImage = img
	} else {
		return fmt.Errorf("no toolbox image configured, the image key is '%s'", controller.ToolboxImage)
	}

	// Add toolbox init container, which will copy necessary tools from cyclone toolbox image to the toolbox
	// emptyDir volume, so that these tools can be shared by other containers. Here we will copy 'fstream' to
	// help collect logs to external log collector (cyclone-server by default if not configured)
	toolbox := corev1.Container{
		Name:  InputContainerName(0),
		Image: toolboxImage,
		Args:  []string{"cp", fmt.Sprintf("%s/fstream", common.ToolboxPath), fmt.Sprintf("%s/resolver-runner.sh", common.ToolboxPath), common.ToolboxVolumeMountPath},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      common.ToolsVolume,
				MountPath: common.ToolboxVolumeMountPath,
			},
		},
		ImagePullPolicy: controller.ImagePullPolicy(),
	}
	m.pod.Spec.InitContainers = append(m.pod.Spec.InitContainers, toolbox)

	for index, r := range m.stg.Spec.Pod.Inputs.Resources {
		log.WithField("stg", m.stage).WithField("resource", r.Name).Debug("Start resolve input resource")
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			log.WithField("resource", r.Name).Error("Get resource error: ", err)
			return err
		}

		// Volume to hold resource data, by default, it's the common PVC in Cyclone, but user can
		// also specify it in the resource spec.
		volumeName := common.DefaultPvVolumeName

		// Sub-path in the PVC to hold resource data
		subPath := common.ResourcePath(m.wfr.Name, r.Name)

		// If persistent is set in the resource spec, create a volume for the persistent PVC
		// specified. Then resource would be pulled in the PVC. If persistent is not set, resource
		// would be pulled in the common PVC in Cyclone. Note that, data in common PVC would be
		// cleaned after workflow terminated.
		persistent := resource.Spec.Persistent
		if persistent != nil {
			subPath = persistent.Path
			volumeName = m.CreatePVCVolume(common.InputResourceVolumeName(r.Name), persistent.PVC)
		} else if m.executionContext.PVC == "" {
			volumeName = GetResourceVolumeName(resource.Name)
			m.CreateEmptyDirVolume(volumeName)
			subPath = ""
		}

		// Create init container for each input resource and project all parameters into the
		// container through environment variables. Parameters are gathered from both the resource
		// spec and the WorkflowRun spec.
		envsMap := make(map[string]string)
		envsMap[common.EnvWorkflowrunName] = m.wfr.Name
		for _, p := range resource.Spec.Parameters {
			if p.Value != nil {
				envsMap[p.Name] = *p.Value
			}
		}

		for _, p := range m.wfr.Spec.Resources {
			if p.Name == r.Name {
				for _, c := range p.Parameters {
					if c.Value != nil {
						envsMap[c.Name] = *c.Value
					}
				}
			}
		}
		var envs []corev1.EnvVar
		for key, value := range envsMap {
			resolved, err := m.refProcessor.ResolveRefStringValue(value, m.client)
			if err != nil {
				return fmt.Errorf("resolve ref value '%s' error: %v", value, err)
			}
			envs = append(envs, corev1.EnvVar{
				Name:  key,
				Value: resolved,
			})
		}

		envs = append(envs, corev1.EnvVar{
			Name: common.EnvLogCollectorURL,
			Value: fmt.Sprintf("%s/apis/v1alpha1/workflowruns/%s/streamlogs?namespace=%s&stage=%s&container=%s",
				controller.Config.CycloneServerAddr, m.wfr.Name, m.wfr.Namespace, m.stg.Name, InputContainerName(index+1)),
		})

		// Get resource resolver for the given resource type. If the resource has resolver set, use it directly,
		// otherwise get resolver from registered resource types.
		resolver, err := common.GetResourceResolver(m.client, resource)
		if err != nil {
			return fmt.Errorf("get resource resolver for resource type '%s' error: %v", resource.Spec.Type, err)
		}

		container := corev1.Container{
			Name:    InputContainerName(index + 1),
			Image:   resolver,
			Command: []string{fmt.Sprintf("%s/resolver-runner.sh", common.ToolboxPath)},
			Args:    []string{common.ResourcePullCommand},
			Env:     envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      volumeName,
					MountPath: common.ResolverDefaultWorkspacePath,
					SubPath:   subPath,
				},
				{
					Name:      common.ToolsVolume,
					MountPath: common.ToolboxPath,
				},
			},
			ImagePullPolicy: controller.ImagePullPolicy(),
		}
		m.pod.Spec.InitContainers = append(m.pod.Spec.InitContainers, container)

		// Mount the resource to all workload containers.
		var containers []corev1.Container
		for _, c := range m.pod.Spec.Containers {
			tmpSubPath := subPath
			if tmpSubPath == "" {
				tmpSubPath = "data"
			} else {
				tmpSubPath = tmpSubPath + string(os.PathSeparator) + "data"
			}

			// We only mount resource to workload containers, sidecars are excluded.
			if common.OnlyWorkload(c.Name) {
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      volumeName,
					MountPath: r.Path,
					SubPath:   tmpSubPath,
				})
			}
			containers = append(containers, c)
		}
		m.pod.Spec.Containers = containers
	}

	return nil
}

// ResolveOutputResources add resource resolvers to pod spec.
func (m *Builder) ResolveOutputResources() error {
	// Indicate whether there is image type resource to output, if so, we need a docker-in-docker
	// side-car.
	var withImageOutput bool

	for index, r := range m.stg.Spec.Pod.Outputs.Resources {
		log.WithField("stg", m.stage).WithField("resource", r.Name).Debug("Start resolve output resource")
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			log.WithField("resource", r.Name).Error("Get resource error: ", err)
			return err
		}

		m.outputResources = append(m.outputResources, resource)

		// Create container for each output resource and project all parameters into the
		// container through environment variables.
		envsMap := make(map[string]string)
		for _, p := range resource.Spec.Parameters {
			if p.Value != nil {
				envsMap[p.Name] = *p.Value
			}
		}
		for _, p := range m.wfr.Spec.Resources {
			if p.Name == r.Name {
				for _, c := range p.Parameters {
					if c.Value != nil {
						envsMap[c.Name] = *c.Value
					}
				}
			}
		}
		var envs []corev1.EnvVar
		for key, value := range envsMap {
			resolved, err := m.refProcessor.ResolveRefStringValue(value, m.client)
			if err != nil {
				return fmt.Errorf("resolve ref value '%s' error: %v", value, err)
			}
			envs = append(envs, corev1.EnvVar{
				Name:  key,
				Value: resolved,
			})
		}

		// Get resource resolver for the given resource type. If the resource has resolver set, use it directly,
		// otherwise get resolver from registered resource types.
		resolver, err := common.GetResourceResolver(m.client, resource)
		if err != nil {
			return fmt.Errorf("get resource resolver for resource type '%s' error: %v", resource.Spec.Type, err)
		}

		container := corev1.Container{
			Name:  OutputContainerName(index + 1),
			Image: resolver,
			Args:  []string{common.ResourcePushCommand},
			Env:   envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      common.CoordinatorSidecarVolumeName,
					MountPath: common.ResolverNotifyDirPath,
					SubPath:   common.ResolverNotifyDir,
				},
			},
			ImagePullPolicy: controller.ImagePullPolicy(),
		}

		if resource.Spec.Persistent != nil {
			volumeName := m.CreatePVCVolume(common.OutputResourceVolumeName(r.Name), resource.Spec.Persistent.PVC)
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      volumeName,
				MountPath: filepath.Join(common.ResolverDefaultDataPath, filepath.Base(resource.Spec.Persistent.Path)),
				SubPath:   resource.Spec.Persistent.Path,
			})
		} else {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      common.CoordinatorSidecarVolumeName,
				MountPath: common.ResolverDefaultDataPath,
				SubPath:   fmt.Sprintf("resources/%s", resource.Name),
			})
		}

		if resource.Spec.Type == v1alpha1.ImageResourceType {
			withImageOutput = true

			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      common.DockerInDockerSockVolume,
				MountPath: common.DockerSockPath,
			})
		}

		m.pod.Spec.Containers = append(m.pod.Spec.Containers, container)
	}

	// Add a volume for docker socket file sharing if there are image type resource to output.
	if withImageOutput {
		m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
			Name: common.DockerInDockerSockVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// Add a docker-in-docker sidecar when there are image type resource to output.
	args := []string{"dockerd"}
	if len(controller.Config.DindSettings.Bip) > 0 {
		args = append(args, "--bip", controller.Config.DindSettings.Bip)
	}
	for _, r := range controller.Config.DindSettings.InsecureRegistries {
		args = append(args, "--insecure-registry", r)
	}

	if withImageOutput {
		var previleged = true
		dind := corev1.Container{
			Image: controller.Config.Images[controller.DindImage],
			Name:  common.DockerInDockerSidecarName,
			Args:  args,
			SecurityContext: &corev1.SecurityContext{
				Privileged: &previleged,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      common.DockerInDockerSockVolume,
					MountPath: common.DockerSockPath,
				},
			},
		}
		m.pod.Spec.Containers = append(m.pod.Spec.Containers, dind)
	}

	// Mount docker socket file to workload container if there are image type resource to output.
	if withImageOutput {
		for i, c := range m.pod.Spec.Containers {
			if common.OnlyCustomContainer(c.Name) {
				m.pod.Spec.Containers[i].VolumeMounts = append(m.pod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
					Name:      common.DockerInDockerSockVolume,
					MountPath: common.DockerSockPath,
				})
			}
		}
	}

	return nil
}

// ResolveInputArtifacts mount each input artifact from PVC.
func (m *Builder) ResolveInputArtifacts() error {
	if m.executionContext.PVC == "" && len(m.stg.Spec.Pod.Inputs.Artifacts) > 0 {
		return fmt.Errorf("artifacts not supported when no PVC provided, but %d input artifacts found", len(m.stg.Spec.Pod.Inputs.Artifacts))
	}

	// Bind input artifacts to workload containers.
	// First find StageItem from Workflow spec, we will get artifacts binding info from it.
	var wfStage *v1alpha1.StageItem
	for _, s := range m.wf.Spec.Stages {
		if s.Name == m.stg.Name {
			wfStage = &s
			break
		}
	}
	if wfStage == nil {
		log.WithField("stg", m.stg.Name).WithField("wfr", m.wf.Name).Error("Stage not found in Workflow")
		return fmt.Errorf("stage %s not found in workflow %s", m.stg.Name, m.wf.Name)
	}

	// For each input artifact, mount data from PVC.
	for _, artifact := range m.stg.Spec.Pod.Inputs.Artifacts {
		// Get source of this input artifact from Workflow StageItem
		// It has format: <stage name>/<artifact name>
		var source string
		for _, art := range wfStage.Artifacts {
			if art.Name == artifact.Name {
				source = art.Source
			}
		}
		if source == "" {
			log.WithField("stg", m.stg.Name).
				WithField("wfr", m.wf.Name).
				WithField("artifact", artifact.Name).
				Error("Input artifact not bind in workflow")
			return fmt.Errorf("input artifact %s not binded in workflow %s", m.stg.Name, m.wf.Name)
		}
		parts := strings.Split(source, "/")
		log.WithField("source", source).
			WithField("artifact", artifact.Name).
			Info("To mount artifact")

		// Mount artifacts to each workload container.
		var containers []corev1.Container
		for _, c := range m.pod.Spec.Containers {
			fileName, err := m.ArtifactFileName(parts[0], parts[1])
			if err != nil {
				return err
			}

			// Mount artifacts only to workload containers, with sidecars excluded.
			if common.OnlyWorkload(c.Name) {
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      common.DefaultPvVolumeName,
					MountPath: artifact.Path,
					SubPath:   common.ArtifactPath(m.wfr.Name, parts[0], parts[1]) + "/" + fileName,
				})
			}
			containers = append(containers, c)
		}
		m.pod.Spec.Containers = containers
	}

	return nil
}

// AddVolumeMounts add common PVC  to workload containers
func (m *Builder) AddVolumeMounts() error {
	if m.executionContext.PVC != "" {
		var containers []corev1.Container
		for _, c := range m.pod.Spec.Containers {
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      common.DefaultPvVolumeName,
				MountPath: common.StageMountPath,
				SubPath:   common.StagePath(m.wfr.Name, m.stg.Name),
			})
			containers = append(containers, c)
		}
		m.pod.Spec.Containers = containers
	}

	return nil
}

// MountPresetVolumes add preset volumes defined in WorkflowRun.
func (m *Builder) MountPresetVolumes() error {
	var containers []corev1.Container
	for _, c := range m.pod.Spec.Containers {
		for i, v := range m.wfr.Spec.PresetVolumes {
			if !MatchContainerGroup(v.ContainerGroup, c.Name) {
				continue
			}

			switch v.Type {
			case v1alpha1.PresetVolumeTypeHostPath:
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      common.PresetVolumeName(i),
					MountPath: v.MountPath,
					ReadOnly:  true,
				})
			case v1alpha1.PresetVolumeTypePVC:
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      common.DefaultPvVolumeName,
					MountPath: v.MountPath,
					SubPath:   v.Path,
				})
			case v1alpha1.PresetVolumeTypeSecret, v1alpha1.PresetVolumeTypeConfigMap:
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      common.PresetVolumeName(i),
					MountPath: v.MountPath,
				})
			default:
				log.WithField("type", v.Type).Warning("Common volume type not supported")
			}
		}
		containers = append(containers, c)
	}

	m.pod.Spec.Containers = containers
	return nil
}

// AddCoordinator adds coordinator container as sidecar to pod. Coordinator is used
// to collect logs, artifacts and notify resource resolvers to push resources.
func (m *Builder) AddCoordinator() error {
	// Get workload container name, for the moment, we support only one workload container.
	var workloadContainer string
	for _, c := range m.stg.Spec.Pod.Spec.Containers {
		workloadContainer = c.Name
		break
	}

	newStg := m.stg.DeepCopy()
	newStg.Spec = *m.rendered
	stgInfo, err := json.Marshal(newStg)
	if err != nil {
		log.Errorf("Marshal stage %s error %s", m.stg.Name, err)
		return err
	}

	wfrInfo, err := json.Marshal(m.wfr)
	if err != nil {
		log.Errorf("Marshal workflowrun %s error %s", m.wfr.Name, err)
		return err
	}

	rscInfo, err := json.Marshal(m.outputResources)
	if err != nil {
		log.Errorf("Marshal output resources error %s", err)
		return err
	}

	coordinator := corev1.Container{
		Name:  common.CoordinatorSidecarName,
		Image: controller.Config.Images[controller.CoordinatorImage],
		Env: []corev1.EnvVar{
			{
				Name:  common.EnvStagePodName,
				Value: m.pod.Name,
			},
			{
				Name:  common.EnvNamespace,
				Value: m.executionContext.Namespace,
			},
			{
				Name:  common.EnvWorkloadContainerName,
				Value: workloadContainer,
			},
			{
				Name:  common.EnvCycloneServerAddr,
				Value: controller.Config.CycloneServerAddr,
			},
			{
				Name:  common.EnvStageInfo,
				Value: string(stgInfo),
			},
			{
				Name:  common.EnvWorkflowRunInfo,
				Value: string(wfrInfo),
			},
			{
				Name:  common.EnvOutputResourcesInfo,
				Value: string(rscInfo),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      common.CoordinatorSidecarVolumeName,
				MountPath: common.CoordinatorResolverPath,
			},
			{
				Name:      common.HostDockerSockVolumeName,
				MountPath: common.DockerSockFilePath,
			},
		},
		ImagePullPolicy: controller.ImagePullPolicy(),
	}
	if m.executionContext.PVC != "" {
		coordinator.VolumeMounts = append(coordinator.VolumeMounts, corev1.VolumeMount{
			Name:      common.DefaultPvVolumeName,
			MountPath: common.CoordinatorWorkspacePath + "artifacts",
			SubPath:   common.ArtifactsPath(m.wfr.Name, m.stage),
		})
	}
	m.pod.Spec.Containers = append(m.pod.Spec.Containers, coordinator)

	return nil
}

// InjectEnvs injects environment variables to containers, such as WorkflowRun name
// stage name, namespace.
func (m *Builder) InjectEnvs() error {
	envs := []corev1.EnvVar{
		{
			Name:  common.EnvWorkflowrunName,
			Value: m.wfr.Name,
		},
		{
			Name:  common.EnvStageName,
			Value: m.stage,
		},
	}
	var containers []corev1.Container
	for _, c := range m.pod.Spec.Containers {
		c.Env = append(c.Env, envs...)
		containers = append(containers, c)
	}
	m.pod.Spec.Containers = containers

	return nil
}

// divideResourceRequirements divides resources requirements by n
func divideResourceRequirements(resources corev1.ResourceRequirements, n int) (*corev1.ResourceRequirements, error) {
	if n <= 0 {
		return &resources, fmt.Errorf("containers count %d can not equal or less than 0", n)
	}

	requirements := resources.DeepCopy()
	for k, v := range resources.Requests {
		newValue, err := divideQuantity(v, n)
		if err != nil {
			return requirements, err
		}
		requirements.Requests[k] = newValue
	}

	for k, v := range resources.Limits {
		newValue, err := divideQuantity(v, n)
		if err != nil {
			return requirements, err
		}
		requirements.Limits[k] = newValue
	}

	return requirements, nil
}

// divideQuantity divides resource.Quantity by n
func divideQuantity(quantity resource.Quantity, n int) (resource.Quantity, error) {
	var newValue resource.Quantity
	var parser = regexp.MustCompile(`([\d.]+)([a-zA-Z]*)`)
	matches := parser.FindStringSubmatch(quantity.String())

	var digit, unit string
	if len(matches) < 2 {
		return newValue, fmt.Errorf("parse Quantity %s error", quantity.String())
	}

	floatValue, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return newValue, err
	}
	digit = strconv.FormatFloat(floatValue/float64(n), 'f', -1, 64)

	if len(matches) > 2 {
		unit = matches[2]
	}

	newValue, err = resource.ParseQuantity(digit + unit)
	if err != nil {
		return newValue, err
	}

	return newValue, nil
}

// applyResourceRequirements applies resource requirements to two types of containers.
// - cyclone containers will be assigned resource requirements with 0(unbounded), since they only
//   consume negligible resources.
// - the other containers(workload containers or custom containers, there is only one at most of the time)
//   will average the pod resource requirements
func applyResourceRequirements(containers []corev1.Container, requirements *corev1.ResourceRequirements, averageToContainers bool) []corev1.Container {
	var results []corev1.Container

	newRequirements := requirements.DeepCopy()
	if averageToContainers {
		containerCount := 0
		for _, c := range containers {
			if common.OnlyCustomContainer(c.Name) {
				containerCount++
			}
		}

		newRequirements, _ = divideResourceRequirements(*requirements, containerCount)
	}

	for _, c := range containers {
		// Set resource requests if not set in the container yet.
		for k, v := range newRequirements.Requests {
			if c.Resources.Requests == nil {
				c.Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
			}

			if _, ok := c.Resources.Requests[k]; !ok {
				if common.OnlyCustomContainer(c.Name) {
					c.Resources.Requests[k] = v
				} else {
					c.Resources.Requests[k] = zeroQuantity
				}
			}

		}

		// Set resource limits if not set in the container yet.
		for k, v := range newRequirements.Limits {
			if c.Resources.Limits == nil {
				c.Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
			}

			if _, ok := c.Resources.Limits[k]; !ok {
				if common.OnlyCustomContainer(c.Name) {
					c.Resources.Limits[k] = v
				} else {
					c.Resources.Limits[k] = zeroQuantity
				}
			}
		}

		results = append(results, c)
	}

	return results
}

// ApplyResourceRequirements applies resource requirements containers in the pod. Resource requirements can be specified
// in three places (ordered by priority descending order):
// - In the Stage spec
// - In the Workflow spec
// - In the Workflow Controller configurations as default values.
// So requirements set in stage spec would have the highest priority.
func (m *Builder) ApplyResourceRequirements() error {
	requirements := &controller.Config.ResourceRequirements
	if m.wf.Spec.Resources != nil {
		requirements = m.wf.Spec.Resources
	}

	m.pod.Spec.InitContainers = applyResourceRequirements(m.pod.Spec.InitContainers, requirements, false)
	m.pod.Spec.Containers = applyResourceRequirements(m.pod.Spec.Containers, requirements, true)

	return nil
}

// ApplyServiceAccount applies service account to pod
func (m *Builder) ApplyServiceAccount() error {
	// If a service account has been explicitly set in stage spec, use it.
	if m.pod.Spec.ServiceAccountName != "" {
		return nil
	}

	if len(controller.Config.ExecutionContext.ServiceAccount) != 0 {
		m.pod.Spec.ServiceAccountName = controller.Config.ExecutionContext.ServiceAccount
	} else {
		m.pod.Spec.ServiceAccountName = common.DefaultServiceAccountName
	}

	return nil
}

// Build ...
func (m *Builder) Build() (*corev1.Pod, error) {
	err := m.Prepare()
	if err != nil {
		return nil, err
	}

	err = m.EnsureContainerNames()
	if err != nil {
		return nil, err
	}

	err = m.ResolveArguments()
	if err != nil {
		return nil, err
	}

	err = m.CreateVolumes()
	if err != nil {
		return nil, err
	}

	err = m.ResolveInputResources()
	if err != nil {
		return nil, err
	}

	err = m.ResolveOutputResources()
	if err != nil {
		return nil, err
	}

	err = m.ResolveInputArtifacts()
	if err != nil {
		return nil, err
	}

	err = m.AddVolumeMounts()
	if err != nil {
		return nil, err
	}

	err = m.MountPresetVolumes()
	if err != nil {
		return nil, err
	}

	err = m.AddCoordinator()
	if err != nil {
		return nil, err
	}

	err = m.InjectEnvs()
	if err != nil {
		return nil, err
	}

	err = m.ApplyResourceRequirements()
	if err != nil {
		return nil, err
	}

	err = m.ApplyServiceAccount()
	if err != nil {
		return nil, err
	}

	return m.pod, nil
}

// ArtifactFileName gets artifact file name from artifacts path.
func (m *Builder) ArtifactFileName(stageName, artifactName string) (string, error) {
	stage, err := m.client.CycloneV1alpha1().Stages(m.wfr.Namespace).Get(stageName, metav1.GetOptions{})
	if err != nil {
		log.WithField("stg", stageName).Error("Get stage error: ", err)
		return "", err
	}

	for _, artifact := range stage.Spec.Pod.Outputs.Artifacts {
		if artifact.Name == artifactName {
			parts := strings.Split(strings.TrimSuffix(artifact.Path, "/"), "/")
			return parts[len(parts)-1], nil
		}
	}

	return "", fmt.Errorf("output artifact '%s' not found in stage '%s'", artifactName, stageName)
}
