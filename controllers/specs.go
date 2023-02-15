package controllers

import (
	"math/rand"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const numOfReplicas = 1
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$&*"

func (r *PgDeployerReconciler) buildDeployment(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*appsv1.Deployment, error) {
	component := "pgdeployment"
	replicas := int32(numOfReplicas)
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: buildMetadata(ns, component),
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(component),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: buildMetadata(ns, component),
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
								{
									Name:  "POSTGRES_DATABASE",
									Value: "postgres",
								},
								{
									Name:  "POSTGRES_USER",
									Value: "postgres",
								},
								getEnvVarSecretSource("POSTGRES_PASSWORD", "pg-secret", "postgres_password"),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	if !testing {
		if err := controllerutil.SetControllerReference(pgDeploy, deploy, r.Scheme); err != nil {
			return nil, err
		}
	}
	return deploy, nil
}

// buildConfigMap will build a kubernetes config map for postgres
func (r *PgDeployerReconciler) buildConfigMap(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*v1.ConfigMap, error) {
	component := "pg-cm"
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "batch/v1/beta1",
		},
		ObjectMeta: buildMetadata(ns, component),
		Data: map[string]string{
			"pg_hba.conf":     "###",
			"postgresql.conf": "data_directory = /var/lib/postgresql/data/data-directory",
		},
	}

	if !testing {
		if err := controllerutil.SetControllerReference(pgDeploy, cm, r.Scheme); err != nil {
			return nil, err
		}
	}

	return cm, nil
}

// buildSecret kubenretes secret for postgres (password generated on the fly and to get it use kubectl get sercet secret-name -o yaml etc.)
func (r *PgDeployerReconciler) buildSecret(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*v1.Secret, error) {
	component := "pg-secret"
	sec := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component),
		StringData: createSecret(),
	}

	if err := controllerutil.SetControllerReference(pgDeploy, sec, r.Scheme); err != nil {
		return nil, err
	}

	return sec, nil
}

// buildPersistentVolume in kubernetes
func (r *PgDeployerReconciler) buildPersistentVolume(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*v1.PersistentVolume, error) {
	component := "pg-pv"
	pv := &v1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component),
		Spec:       v1.PersistentVolumeSpec{},
	}

	if err := controllerutil.SetControllerReference(pgDeploy, pv, r.Scheme); err != nil {
		return nil, err
	}

	return pv, nil
}

// buildPersistentVolumeClaim from the persistent volume
func (r *PgDeployerReconciler) buildPersistentVolumeClaim(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*v1.PersistentVolumeClaim, error) {
	component := "pg-pvc"
	pvc := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component),
		Spec:       v1.PersistentVolumeClaimSpec{},
	}

	if err := controllerutil.SetControllerReference(pgDeploy, pvc, r.Scheme); err != nil {
		return nil, err
	}

	return pvc, nil
}

// buildService in kubernetes with pgDeploy port
func (r *PgDeployerReconciler) buildService(ns string, testing bool, pgDeploy *v1alpha1.PgDeployer) (*v1.Service, error) {
	component := "pg-service"
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: buildMetadata(ns, component),
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

	if err := controllerutil.SetControllerReference(pgDeploy, svc, r.Scheme); err != nil {
		return nil, err
	}

	return svc, nil
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

func buildMetadata(ns, component string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      component,
		Namespace: ns,
		Labels:    buildLabels(component),
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
