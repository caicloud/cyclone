package cd

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// UpdateDeployment updates images in a deployment.
func UpdateDeployment(client kubernetes.Interface, config *Config) error {
	if config.Deployment.Type != DeploymentTypeDeployment {
		return fmt.Errorf("expect type '%s', but got %s", DeploymentTypeDeployment, config.Deployment.Type)
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deploy, err := client.AppsV1().Deployments(config.Deployment.Namespace).Get(context.TODO(), config.Deployment.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		var containers []corev1.Container
		for i, c := range deploy.Spec.Template.Spec.Containers {
			for _, u := range config.Images {
				if u.Container == fmt.Sprintf("#%d", i) || u.Container == c.Name {
					c.Image = u.Image
					break
				}
			}
			containers = append(containers, c)
		}

		deploy.Spec.Template.Spec.Containers = containers
		_, err = client.AppsV1().Deployments(config.Deployment.Namespace).Update(context.TODO(), deploy, metav1.UpdateOptions{})
		return err
	})
}
