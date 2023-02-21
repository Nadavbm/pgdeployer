package v1alpha1

import (
	"math/rand"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const numOfReplicas = 1
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$&*"

// ConstructObjectsFromSpecifications construct a slice of kubernetes object interfaces from specifications
func (pd *PgDeployer) ConstructObjectsFromSpecifications(ns string) []metav1.Object {
	var objects []metav1.Object

	cm := buildConfigMap(ns, pd)
	secret := buildSecret(ns, pd)
	deploy := buildDeployment(ns, pd)
	svc := buildService(ns, pd)

	objects = append(objects, cm, secret, deploy, svc)

	return objects
}

// buildDeployment creates a kubernetes deployment specification
func buildDeployment(ns string, pgDeploy *PgDeployer) *appsv1.Deployment {
	component := "pgdeployment"
	replicas := int32(numOfReplicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: buildMetadata(ns, component, pgDeploy),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(component),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: buildMetadata(ns, component, pgDeploy),
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "database",
							Image: "postgres:" + pgDeploy.Spec.PgVersion,
							Ports: []v1.ContainerPort{
								{
									Protocol:      v1.ProtocolTCP,
									ContainerPort: pgDeploy.Spec.ContainerPort,
								},
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(pgDeploy.Spec.MemoryLimit),
									v1.ResourceCPU:    resource.MustParse(pgDeploy.Spec.CpuLimit),
								},
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse(pgDeploy.Spec.MemoryRequest),
									v1.ResourceCPU:    resource.MustParse(pgDeploy.Spec.CpuRequest),
								},
							},
							Env: []v1.EnvVar{
								getEnvVarSecretSource("POSTGRES_DATABASE", "pg-secret", "postgres_database"),
								getEnvVarSecretSource("POSTGRES_USER", "pg-secret", "postgres_user"),
								getEnvVarSecretSource("POSTGRES_PASSWORD", "pg-secret", "postgres_password"),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}
}

// buildConfigMap will build a kubernetes config map for postgres
func buildConfigMap(ns string, pgDeploy *PgDeployer) *v1.ConfigMap {
	component := "pg-cm"
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "batch/v1/beta1",
		},
		ObjectMeta: buildMetadata(ns, component, pgDeploy),
		Data: map[string]string{
			"pg_hba.conf":     "###",
			"postgresql.conf": "data_directory = /var/lib/postgresql/data/data-directory",
		},
	}
}

// buildSecret kubenretes secret for postgres (password generated on the fly and to get it use kubectl get sercet secret-name -o yaml etc.)
func buildSecret(ns string, pgDeploy *PgDeployer) *v1.Secret {
	component := "pg-secret"
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component, pgDeploy),
		StringData: createSecret(),
	}
}

// buildService in kubernetes with pgDeploy port
func buildService(ns string, pgDeploy *PgDeployer) *v1.Service {
	component := "pg-service"
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component, pgDeploy),
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port: pgDeploy.Spec.ContainerPort,
				},
			},
			Selector: buildLabels(component),
		},
	}
}

//
// ------------------------------------------------------------------------------------------------------- helpers -----------------------------------------------------------------------------
//

func getEnvVarSecretSource(envName, name, secret string) v1.EnvVar {
	return v1.EnvVar{
		Name: envName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: name,
				},
				Key: secret,
			},
		},
	}
}

func buildMetadata(ns, component string, pgDeploy *PgDeployer) metav1.ObjectMeta {
	controller := true
	return metav1.ObjectMeta{
		Name:      component,
		Namespace: ns,
		Labels:    buildLabels(component),
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: pgDeploy.APIVersion,
				Kind:       pgDeploy.Kind,
				Name:       pgDeploy.Name,
				UID:        pgDeploy.UID,
				Controller: &controller,
			},
		},
	}
}

func buildLabels(component string) map[string]string {
	m := make(map[string]string)
	m["app"] = "postgres"
	m["app.kubernetes.io/name"] = component
	m["app.kubernetes.io/component"] = component
	return m
}

func createSecret() map[string]string {
	m := make(map[string]string)
	m["postgres_database"] = "postgres"
	m["postgres_user"] = "postgres"
	m["postgres_password"] = randStringBytes(12)
	return m
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
