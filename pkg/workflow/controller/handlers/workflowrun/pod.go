package workflowrun

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

type PodBuilder struct {
	client clientset.Interface
	wf     *v1alpha1.Workflow
	wfr    *v1alpha1.WorkflowRun
	stg    *v1alpha1.Stage
	stage  string
	pod    *corev1.Pod
}

func NewPodBuilder(client clientset.Interface, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun, stage string) *PodBuilder {
	return &PodBuilder{
		client: client,
		wf:     wf,
		wfr:    wfr,
		stage:  stage,
		pod:    &corev1.Pod{},
	}
}

func (m *PodBuilder) Prepare() error {
	stage, err := m.client.CycloneV1alpha1().Stages(m.wfr.Namespace).Get(m.stage, metav1.GetOptions{})
	if err != nil {
		return err
	}
	m.stg = stage

	// TODO(ChenDe): Support template.
	if stage.Spec.Template != nil {
		return fmt.Errorf("Stage template not support yet, stage: %s", m.stage)
	}
	if stage.Spec.Pod == nil {
		return fmt.Errorf("pod must be defined in stage spec, stage: %s", m.stage)
	}

	// TODO(ChenDe): Support multiple containers in pod workload.
	if len(stage.Spec.Pod.Spec.Containers) != 1 {
		return fmt.Errorf("only one container in pod spec supported, stage: %s", m.stage)
	}

	// Generate pod name using UUID.
	id := uuid.NewV1()
	if err != nil {
		return err
	}
	podName := fmt.Sprintf("%s-%s-%s", m.wf.Name, m.stage, strings.Replace(id.String(), "-", "", -1))
	m.pod.ObjectMeta = metav1.ObjectMeta{
		Name:      podName,
		Namespace: m.wfr.Namespace,
		Labels: map[string]string{
			common.WorkflowLabelName: "true",
		},
		Annotations: map[string]string{
			common.WorkflowRunAnnotationName: m.wfr.Name,
			common.StageAnnotationName:       m.stage,
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: v1alpha1.APIVersion,
				Kind:       "WorkflowRun",
				Name:       m.wfr.Name,
				UID:        m.wfr.UID,
			},
		},
	}

	return nil
}

// TODO(ChenDe): Implement stage template.
func (m *PodBuilder) ApplyTemplate() error {
	return nil
}

func (m *PodBuilder) ResolveArguments() error {
	parameters := make(map[string]string)
	for _, s := range m.wfr.Spec.Stages {
		if s.Name == m.stage {
			for _, p := range s.Parameters {
				parameters[p.Name] = p.Value
			}
		}
	}
	for _, a := range m.stg.Spec.Pod.Inputs.Arguments {
		if _, ok := parameters[a.Name]; !ok {
			if a.Value == "" {
				log.WithField("Argument", a.Name).
					WithField("Stage", m.stg.Name).
					Error("Argument not set and without default value")
				return fmt.Errorf("argument '%s' not set in stage '%s' and without default value", a.Name, m.stg.Name)
			}
			parameters[a.Name] = a.Value
		}
	}
	log.WithField("Parameters", parameters).Debug("Parameters collected")
	raw, err := json.Marshal(m.stg.Spec.Pod.Spec)
	if err != nil {
		return err
	}
	rendered, err := mustache.Render(string(raw), parameters)
	if err != nil {
		return err
	}
	renderedSpec := corev1.PodSpec{}
	json.Unmarshal([]byte(rendered), &renderedSpec)
	m.pod.Spec = renderedSpec
	m.pod.Spec.RestartPolicy = corev1.RestartPolicyNever

	return nil
}

func (m *PodBuilder) CreateVolumes() error {
	// Add volumes for input resources to pod
	for _, r := range m.stg.Spec.Pod.Inputs.Resources {
		m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
			Name: r.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// Add emptyDir volume to be shared between coordinator and sidecars, e.g. resource resolvers.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.CoordinatorSidecarVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Add PVC volume to pod
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.DefaultPvVolumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: controller.Config.PVC,
			},
		},
	})

	// Create hostPath volume for /var/run/docker.sock
	var hostPathSocket = corev1.HostPathSocket
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.DockerSockVolume,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: common.DockerSockPath,
				Type: &hostPathSocket,
			},
		},
	})

	// Create secret volume for use in resource resolvers.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: common.DockerConfigJsonVolume,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: common.DefaultSecretName,
				Items: []corev1.KeyToPath{
					{
						Key: common.DockerConfigJsonFile,
						Path: common.DockerConfigJsonFile,
					},
				},
			},
		},
	})

	return nil
}

// ResolveInputResources creates init containers for each input resource and also mount
// resource to workload containers.
func (m *PodBuilder) ResolveInputResources() error {
	for _, r := range m.stg.Spec.Pod.Inputs.Resources {
		log.WithField("stage", m.stage).WithField("resource", r.Name).Debug("Start resolve input resource")
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			log.WithField("resource", r.Name).Error("Get resource error: ", err)
			return err
		}

		// Get resource resolver image, if the resource is build-in resource (Git, Image, KV), use
		// the images configured, otherwise use images given in the resource spec.
		var image string
		if key, ok := controller.ResolverImageKeys[resource.Spec.Type]; ok {
			image = controller.Config.Images[key]
		} else {
			image = resource.Spec.Resolver
		}

		// Create init container for each input resource and project all parameters into the
		// container through environment variables.
		var envs []corev1.EnvVar
		for _, p := range resource.Spec.Parameters {
			envs = append(envs, corev1.EnvVar{
				Name:  p.Name,
				Value: p.Value,
			})
		}
		container := corev1.Container{
			Name:  r.Name,
			Image: image,
			Args:  []string{common.ResourcePullCommand},
			Env:   envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      r.Name,
					MountPath: common.ResolverDefaultDataPath,
				},
			},
		}
		m.pod.Spec.InitContainers = append(m.pod.Spec.InitContainers, container)

		// Mount the resource to all workload containers.
		var containers []corev1.Container
		for _, c := range m.pod.Spec.Containers {
			// We only mount resource to workload containers, sidecars are excluded.
			if common.OnlyWorkload(c.Name) {
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      r.Name,
					MountPath: r.Path,
				})
			}
			containers = append(containers, c)
		}
		m.pod.Spec.Containers = containers
	}

	return nil
}

// ResolveOutputResources add resource resolvers to pod spec.
func (m *PodBuilder) ResolveOutputResources() error {
	for _, r := range m.stg.Spec.Pod.Outputs.Resources {
		log.WithField("stage", m.stage).WithField("resource", r.Name).Debug("Start resolve output resource")
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			log.WithField("resource", r.Name).Error("Get resource error: ", err)
			return err
		}

		// Get resource resolver image, if the resource is build-in resource (Git, Image, KV), use
		// the images configured, otherwise use images given in the resource spec.
		var image string
		if key, ok := controller.ResolverImageKeys[resource.Spec.Type]; ok {
			image = controller.Config.Images[key]
		} else {
			image = resource.Spec.Resolver
		}

		// Create container for each output resource and project all parameters into the
		// container through environment variables. Also mount volumes shared between resolver
		// and coordinator container.
		var envs []corev1.EnvVar
		for _, p := range resource.Spec.Parameters {
			envs = append(envs, corev1.EnvVar{
				Name:  p.Name,
				Value: p.Value,
			})
		}
		container := corev1.Container{
			Name:  common.CycloneSidecarPrefix + r.Name,
			Image: image,
			Args:  []string{common.ResourcePushCommand},
			Env:   envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      common.CoordinatorSidecarVolumeName,
					MountPath: common.ResolverNotifyDirPath,
					SubPath:   common.ResolverNotifyDir,
				},
				{
					Name:      common.CoordinatorSidecarVolumeName,
					MountPath: common.ResolverDefaultDataPath,
					SubPath:   fmt.Sprintf("resources/%s", resource.Name),
				},
			},
			// TODO(ChenDe): Used for develop purpose only, remove it.
			ImagePullPolicy: corev1.PullAlways,
		}

		if resource.Spec.Type == v1alpha1.ImageResourceType {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name: common.DockerSockVolume,
				MountPath: common.DockerSockPath,
			}, corev1.VolumeMount{
				Name: common.DockerConfigJsonVolume,
				MountPath: common.DockerConfigPath,
			})
		}

		m.pod.Spec.Containers = append(m.pod.Spec.Containers, container)
	}

	return nil
}

// ResolveInputArtifacts mount each input artifact from PVC.
func (m *PodBuilder) ResolveInputArtifacts() error {
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
		log.WithField("stage", m.stg.Name).WithField("workflow", m.wf.Name).Error("Stage not found in Workflow")
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
			log.WithField("stage", m.stg.Name).
				WithField("workflow", m.wf.Name).
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

// AddVolumeMounts add common PVC and coordinator emptyDir volume to workload containers
func (m *PodBuilder) AddVolumeMounts() error {
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

	return nil
}

// AddCoordinator adds coordinator container as sidecar to pod. Coordinator is used
// to collect logs, artifacts and notify resource resolvers to push resources.
func (m *PodBuilder) AddCoordinator() error {
	// Get workload container name, for the moment, we support only one workload container.
	var workloadContainer string
	for _, c := range m.stg.Spec.Pod.Spec.Containers {
		workloadContainer = c.Name
		break
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
				Value: m.wfr.Namespace,
			},
			{
				Name:  common.EnvWorkflowrunName,
				Value: m.wfr.Name,
			},
			{
				Name:  common.EnvStageName,
				Value: m.stage,
			},
			{
				Name:  common.EnvWorkloadContainerName,
				Value: workloadContainer,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      common.DefaultPvVolumeName,
				MountPath: common.CoordinatorWorkspacePath + "artifacts",
				SubPath:   common.ArtifactsPath(m.wfr.Name, m.stage),
			},
			{
				Name:      common.DockerSockVolume,
				MountPath: common.DockerSockPath,
			},
			{
				Name:      common.CoordinatorSidecarVolumeName,
				MountPath: common.CoordinatorResolverPath,
			},
		},
		// TODO(ChenDe): Used for develop purpose only, remove it.
		ImagePullPolicy: corev1.PullAlways,
	}
	m.pod.Spec.Containers = append(m.pod.Spec.Containers, coordinator)

	return nil
}

func (m *PodBuilder) Build() (*corev1.Pod, error) {
	err := m.Prepare()
	if err != nil {
		return nil, err
	}

	err = m.ApplyTemplate()
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

	err = m.AddCoordinator()
	if err != nil {
		return nil, err
	}

	return m.pod, nil
}

func (m *PodBuilder) ArtifactFileName(stageName, artifactName string) (string, error) {
	stage, err := m.client.CycloneV1alpha1().Stages(m.wfr.Namespace).Get(stageName, metav1.GetOptions{})
	if err != nil {
		log.WithField("stage", stageName).Error("Get stage error: ", err)
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
