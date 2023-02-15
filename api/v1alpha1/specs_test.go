package v1alpha1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestKubernetesSpecifications(t *testing.T) {
	var pgDeploy PgDeployer

	deploy := PgDeployerSpec{
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
		t.Errorf("expected cpu limits %v, instead got %v", limits.Cpu(), resourceListLimits.Cpu())
	}

	if limits.Memory().String() != resourceListLimits.Memory().String() {
		t.Errorf("expected memory limits %v, instead got %v", limits.Memory(), resourceListLimits.Memory())
	}

	if requests.Cpu().String() != resourceListRequests.Cpu().String() {
		t.Errorf("expected cpu requests %v, instead got %v", requests.Cpu(), resourceListRequests.Cpu())
	}

	if requests.Memory().String() != resourceListRequests.Memory().String() {
		t.Errorf("expected memory requests %v, instead got %v", requests.Memory(), resourceListRequests.Memory())
	}
}
