package workflowrun

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/workflow"
	"github.com/caicloud/cyclone/pkg/workflow/common"
	"github.com/caicloud/cyclone/pkg/workflow/controller"

	"github.com/cbroglie/mustache"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	return nil
}

func (m *PodBuilder) ApplyTemplate() error {
	if m.stg.Spec.Template != nil {
		// TODO(ChenDe): Implement stage template
		return fmt.Errorf("stage template not supported yet")
	}
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

	// Add emptyDir volume to be shared between coordinator and workload containers.
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: workflow.CoordinatorVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Add PVC volume to pod
	m.pod.Spec.Volumes = append(m.pod.Spec.Volumes, corev1.Volume{
		Name: workflow.DefaultPvVolumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: controller.Config.PVC,
			},
		},
	})

	return nil
}

// ResolveInputResource creates init containers for each input resource and also mount
// resource to workload containers.
func (m *PodBuilder) ResolveInputResource() error {
	for _, r := range m.stg.Spec.Pod.Inputs.Resources {
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
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
			Args:  []string{workflow.ResourcePullCommand},
			Env:   envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      r.Name,
					MountPath: workflow.ResolverDataPath,
				},
			},
		}
		m.pod.Spec.InitContainers = append(m.pod.Spec.InitContainers, container)

		// Mount the resource to all workload containers.
		var containers []corev1.Container
		for _, c := range m.pod.Spec.Containers {
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      r.Name,
				MountPath: r.Path,
			})
			containers = append(containers, c)
		}
		m.pod.Spec.Containers = containers
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
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      workflow.DefaultPvVolumeName,
				MountPath: artifact.Path,
				SubPath:   common.ArtifactPath(m.wfr.Name, parts[0], parts[1]) + "/" + fileName,
			})
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
			Name:      workflow.DefaultPvVolumeName,
			MountPath: common.StageMountPath,
			SubPath:   common.StagePath(m.wfr.Name, m.stg.Name),
		}, corev1.VolumeMount{
			Name:      workflow.CoordinatorVolumeName,
			MountPath: common.StageEmptyDirMounthPath,
		})
		containers = append(containers, c)
	}
	m.pod.Spec.Containers = containers

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

	err = m.ResolveInputResource()
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

	// Generate pod name using based on UUID.
	id := uuid.NewV1()
	if err != nil {
		return nil, err
	}
	podName := fmt.Sprintf("%s-%s-%s", m.wf.Name, m.stage, strings.Replace(id.String(), "-", "", -1))

	// Add coordinator container to containers.
	/*
		coordinator := corev1.Container{
			Name: workflow.SidecarContainerPrefix + workflow.CoordinatorContainerName,
			Image: controller.Config.Images[controller.CoordinatorImage],
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					Value: podName,
				},
				{
					Name: "NAMESPACE",
					Value: wfr.Namespace,
				},
				{
					Name: "WORKFLOWRUN_NAME",
					Value: wfr.Name,
				},
				{
					Name: "STAGE_NAME",
					Value: stageName,
				},
				{
					Name: "WORKLOAD_CONTAINER",
					Value: "",
				},
			},
		}
		renderedSpec.Containers = append(renderedSpec.Containers, coordinator)*/

	m.pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	m.pod.ObjectMeta = metav1.ObjectMeta{
		Name:      podName,
		Namespace: m.wfr.Namespace,
		Labels: map[string]string{
			workflow.WorkflowLabelName: "true",
		},
		Annotations: map[string]string{
			workflow.WorkflowRunAnnotationName: m.wfr.Name,
			workflow.StageAnnotationName:       m.stage,
		},
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
