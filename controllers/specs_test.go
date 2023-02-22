package controllers

import (
	"testing"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestKubernetesSpecifications(t *testing.T) {
	var pgDeploy v1alpha1.PgDeployer

	deploy := v1alpha1.PgDeployerSpec{
		PgVersion:     "14",
		ContainerPort: 5432,
		CpuRequest:    "500m",
		CpuLimit:      "1000m",
		MemoryRequest: "512Mi",
		MemoryLimit:   "1024Mi",
	}
	pgDeploy.Spec = deploy

	resourceListLimits := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse(deploy.MemoryLimit),
		v1.ResourceCPU:    resource.MustParse(deploy.CpuLimit),
	}

	resourceListRequests := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse(deploy.MemoryRequest),
		v1.ResourceCPU:    resource.MustParse(deploy.CpuRequest),
	}

	deployment := buildDeployment("testing", &pgDeploy)

	containers := deployment.Spec.Template.Spec.Containers

	requests := containers[0].Resources.Requests
	limits := containers[0].Resources.Limits

	if limits.Cpu().String() != resourceListLimits.Cpu().String() {
		t.Errorf("expected cpu limits %v, instead got %v", resourceListLimits.Cpu(), limits.Cpu())
	}

	if limits.Memory().String() != resourceListLimits.Memory().String() {
		t.Errorf("expected memory limits %v, instead got %v", resourceListLimits.Memory(), limits.Memory())
	}

	if requests.Cpu().String() != resourceListRequests.Cpu().String() {
		t.Errorf("expected cpu requests %v, instead got %v", resourceListRequests.Cpu(), requests.Cpu())
	}

	if requests.Memory().String() != resourceListRequests.Memory().String() {
		t.Errorf("expected memory requests %v, instead got %v", resourceListRequests.Memory(), requests.Memory())
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != "postgres:14" {
		t.Errorf("expected image to be postgres:14, instead got %s", deployment.Spec.Template.Spec.Containers[0].Image)
	}

	service := buildService("testing", &pgDeploy)

	if service.Spec.Ports[0].Port != pgDeploy.Spec.ContainerPort {
		t.Errorf("expected service port to be %d, instead got %d", pgDeploy.Spec.ContainerPort, service.Spec.Ports[0].Port)
	}

	deploy = v1alpha1.PgDeployerSpec{
		PgVersion:     "13",
		ContainerPort: 5433,
		CpuRequest:    "100m",
		CpuLimit:      "200m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "256Mi",
	}
	pgDeploy.Spec = deploy

	deployment = buildDeployment("testing", &pgDeploy)

	if pgDeploy.Spec.CpuLimit != "200m" {
		t.Errorf("expected cpu limits to be 200m, instead got %v", pgDeploy.Spec.CpuLimit)
	}

	if pgDeploy.Spec.MemoryLimit != "256Mi" {
		t.Errorf("expected memory limits 256Mi, instead got %v", pgDeploy.Spec.MemoryLimit)
	}

	if pgDeploy.Spec.CpuRequest != "100m" {
		t.Errorf("expected cpu requests 100m, instead got %v", pgDeploy.Spec.CpuRequest)
	}

	if pgDeploy.Spec.MemoryRequest != "128Mi" {
		t.Errorf("expected memory requests 128Mi, instead got %v", resourceListRequests.Memory())
	}

	if deployment.Spec.Template.Spec.Containers[0].Image != "postgres:13" {
		t.Errorf("expected image to be postgres:13, instead got %s", deployment.Spec.Template.Spec.Containers[0].Image)
	}

	service = buildService("testing", &pgDeploy)

	if service.Spec.Ports[0].Port != 5433 {
		t.Errorf("expected service port to be 5433, instead got %d", service.Spec.Ports[0].Port)
	}
}
