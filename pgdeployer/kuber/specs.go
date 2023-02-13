package kuber

import (
	"math/rand"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const numOfReplicas = 1
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$&*"

func buildDeployment(ns string, pgDeploy v1alpha1.PgDeployerSpec) *appsv1.Deployment {
	replicas := int32(numOfReplicas)
	specs := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: buildMetadata(ns),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: buildMetadata(ns),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "database",
							Image: "postgres:14",
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: pgDeploy.ContainerPort,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(pgDeploy.MemoryLimit),
									v1.ResourceCPU:    resource.MustParse(pgDeploy.CpuLimit),
								},
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(pgDeploy.MemoryRequest),
									v1.ResourceCPU:    resource.MustParse(pgDeploy.CpuRequest),
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	return specs
}

// buildConfigMap will build a kubernetes config map for postgres
func buildConfigMap(ns string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "batch/v1/beta1",
		},
		ObjectMeta: buildMetadata(ns),
		Data: map[string]string{
			"pg_hba.conf":     "###",
			"postgresql.conf": "data_directory = /var/lib/postgresql/data/data-directory",
		},
	}
	return cm
}

// buildSecret kubenretes secret for postgres (password generated on the fly and to get it use kubectl get sercet secret-name -o yaml etc.)
func buildSecret(ns string) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns),
		StringData: createSecret(),
	}
}

// buildPersistentVolume in kubernetes
func buildPersistentVolume(ns string) *v1.PersistentVolume {
	return &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns),
		Spec:       v1.PersistentVolumeSpec{},
	}
}

// buildPersistentVolumeClaim from the persistent volume
func buildPersistentVolumeClaim(ns string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns),
		Spec:       v1.PersistentVolumeClaimSpec{},
	}
}

// buildService in kubernetes with pgDeploy port
func buildService(ns string, pgDeploy v1alpha1.PgDeployerSpec) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port: pgDeploy.ContainerPort,
				},
			},
			Selector: buildLabels(),
		},
	}
}

//
// ------------------------------------------------------------------------------------------------------- helpers -----------------------------------------------------------------------------
//

func buildMetadata(ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      ns,
		Namespace: ns,
		Labels:    buildLabels(),
	}
}

func buildLabels() map[string]string {
	m := make(map[string]string)
	m["app"] = "postgres"
	return m
}

func createSecret() map[string]string {
	m := make(map[string]string)
	m["POSTGRES_DB"] = "postgres"
	m["POSTGRES_USER"] = "postgres"
	m["POSTGRES_PASSWORD"] = randStringBytes(12)
	return m
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
