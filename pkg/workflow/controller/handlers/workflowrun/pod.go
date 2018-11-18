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

type PodMaker struct {
	client clientset.Interface
	wf     *v1alpha1.Workflow
	wfr    *v1alpha1.WorkflowRun
	stage  string
}

func NewPodMaker(client clientset.Interface, wf *v1alpha1.Workflow, wfr *v1alpha1.WorkflowRun, stage string) *PodMaker {
	return &PodMaker{
		client: client,
		wf:     wf,
		wfr:    wfr,
		stage:  stage,
	}
}

func (m *PodMaker) MakePod() (*corev1.Pod, error) {
	stage, err := m.client.CycloneV1alpha1().Stages(m.wfr.Namespace).Get(m.stage, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// This stage is using template
	if stage.Spec.Template != nil {
		// TODO(ChenDe): Implement stage template
		return nil, fmt.Errorf("stage template not supported yet")
	}

	// Apply parameters to pod spec.
	parameters := make(map[string]string)
	for _, s := range m.wfr.Spec.Stages {
		if s.Name == m.stage {
			for _, p := range s.Parameters {
				parameters[p.Name] = p.Value
			}
		}
	}
	for _, a := range stage.Spec.Pod.Inputs.Arguments {
		_, ok := parameters[a.Name]
		if !ok {
			if a.Value == "" {
				log.WithField("Argument", a.Name).
					WithField("Stage", stage.Name).
					Error("Argument not set and without default value")
				return nil, fmt.Errorf("argument '%s' not set in stage '%s' and without default value", a.Name, stage.Name)
			}
			parameters[a.Name] = a.Value
		}
	}
	log.WithField("Parameters", parameters).Debug("Parameters collected")
	raw, err := json.Marshal(stage.Spec.Pod.Spec)
	if err != nil {
		return nil, err
	}
	rendered, err := mustache.Render(string(raw), parameters)
	if err != nil {
		return nil, err
	}
	renderedSpec := corev1.PodSpec{}
	json.Unmarshal([]byte(rendered), &renderedSpec)

	// Add volumes for input resources to pod
	for _, r := range stage.Spec.Pod.Inputs.Resources {
		renderedSpec.Volumes = append(renderedSpec.Volumes, corev1.Volume{
			Name: r.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// Add emptyDir volume to be shared between coordinator and workload containers.
	renderedSpec.Volumes = append(renderedSpec.Volumes, corev1.Volume{
		Name: workflow.CoordinatorVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})

	// Add PVC volume to pod
	renderedSpec.Volumes = append(renderedSpec.Volumes, corev1.Volume{
		Name: workflow.DefaultPvVolumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: controller.Config.PVC,
			},
		},
	})

	// Add init containers for each input resource and also bind resource to workload containers.
	for _, r := range stage.Spec.Pod.Inputs.Resources {
		resource, err := m.client.CycloneV1alpha1().Resources(m.wfr.Namespace).Get(r.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		var image string
		switch resource.Spec.Type {
		case v1alpha1.GitResourceType:
			image = controller.Config.Images[controller.GitResolverImage]
		case v1alpha1.ImageResourceType:
			image = controller.Config.Images[controller.ImageResolverImage]
		case v1alpha1.KVResourceType:
			image = controller.Config.Images[controller.KvResolverImage]
		case v1alpha1.GeneralResourceType:
			image = resource.Spec.Resolver
		}

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
			Args:  []string{"pull"},
			Env:   envs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      r.Name,
					MountPath: workflow.ResolverDataPath,
				},
			},
		}
		renderedSpec.InitContainers = append(renderedSpec.InitContainers, container)

		// Mount the resource to all workload containers.
		var containers []corev1.Container
		for _, c := range renderedSpec.Containers {
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      r.Name,
				MountPath: r.Path,
			})
			containers = append(containers, c)
		}
		renderedSpec.Containers = containers
	}

	// Bind input artifacts to workload containers.
	// 1. First find StageItem from Workflow spec, we will get artifacts binding info from it.
	var wfStage *v1alpha1.StageItem
	for _, s := range m.wf.Spec.Stages {
		if s.Name == stage.Name {
			wfStage = &s
			break
		}
	}
	if wfStage == nil {
		log.WithField("stage", stage.Name).WithField("workflow", m.wf.Name).Error("Stage not found in Workflow")
		return nil, fmt.Errorf("stage %s not found in workflow %s", stage.Name, m.wf.Name)
	}

	// 2. For each input artifact, mount data from PVC.
	for _, artifact := range stage.Spec.Pod.Inputs.Artifacts {
		// Get source of this input artifact from Workflow StageItem
		// It has format: <stage name>/<artifact name>
		var source string
		for _, art := range wfStage.Artifacts {
			if art.Name == artifact.Name {
				source = art.Source
			}
		}
		if source == "" {
			log.WithField("stage", stage.Name).
				WithField("workflow", m.wf.Name).
				WithField("artifact", artifact.Name).
				Error("Input artifact not bind in workflow")
			return nil, fmt.Errorf("input artifact %s not binded in workflow %s", stage.Name, m.wf.Name)
		}
		parts := strings.Split(source, "/")
		log.WithField("source", source).
			WithField("artifact", artifact.Name).
			Info("To mount artifact")

		// Mount artifacts to each workload container.
		var containers []corev1.Container
		for _, c := range renderedSpec.Containers {
			fileName, err := m.ArtifactFileName(parts[0], parts[1])
			if err != nil {
				return nil, err
			}
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      workflow.DefaultPvVolumeName,
				MountPath: artifact.Path,
				SubPath:   common.ArtifactPath(m.wfr.Name, parts[0], parts[1]) + "/" + fileName,
			})
			containers = append(containers, c)
		}
		renderedSpec.Containers = containers
	}

	// Bind common PVC and coordinator emptyDir to workload containers
	var containers []corev1.Container
	for _, c := range renderedSpec.Containers {
		c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
			Name:      workflow.DefaultPvVolumeName,
			MountPath: common.StageMountPath,
			SubPath:   common.StagePath(m.wfr.Name, stage.Name),
		}, corev1.VolumeMount{
			Name:      workflow.CoordinatorVolumeName,
			MountPath: common.StageEmptyDirMounthPath,
		})
		containers = append(containers, c)
	}
	renderedSpec.Containers = containers

	// Generate pod using based on UUID.
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

	renderedSpec.RestartPolicy = corev1.RestartPolicyNever
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: m.wfr.Namespace,
			Labels: map[string]string{
				workflow.WorkflowLabelName: "true",
			},
			Annotations: map[string]string{
				workflow.WorkflowrunAnnotationName: m.wfr.Name,
				workflow.StageAnnotationName:       m.stage,
			},
		},
		Spec: renderedSpec,
	}

	return pod, nil
}

func (m *PodMaker) ArtifactFileName(stageName, artifactName string) (string, error) {
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
