package kuber

import (
	"context"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"

	"github.com/nadavbm/zlog"
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kuber struct {
	logger *zlog.Logger
	client kubernetes.Clientset
}

// New will create a new instance of kuber
func new(logger *zlog.Logger) (*Kuber, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kuber{
		logger: logger,
		client: *k8sClient,
	}, nil
}

func (k *Kuber) deployPostgresToKubernetes(logger *zlog.Logger, ns string, pgDeploy v1alpha1.PgDeployerSpec) error {
	k, err := new(logger)
	if err != nil {
		return err
	}

	if err := k.ApplyConfigMap(ns); err != nil {
		return err
	}

	if err := k.ApplyDeployment(ns, pgDeploy); err != nil {
		return err
	}

	return nil
}

func (k *Kuber) ApplyDeployment(ns string, pgDeploy v1alpha1.PgDeployerSpec) error {
	deploy := buildDeployment(ns, pgDeploy)
	deploymentInterface := k.client.AppsV1().Deployments(ns)

	deployment, err := deploymentInterface.Create(context.TODO(), deploy, metav1.CreateOptions{})
	if err == nil {
		k.logger.Info("deployment created", zap.String("namespace:", ns), zap.Any("deployment:", deployment))
		return nil
	} else if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	deployExists, err := deploymentInterface.Get(context.TODO(), deploy.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	deployExists.Spec = deploy.Spec
	deployment, err = deploymentInterface.Update(context.TODO(), deployExists, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	k.logger.Info("deployment updated", zap.String("namespace:", ns), zap.Any("deployment:", deployment))
	return nil
}

// ApplyConfigMap will apply the configmap in kubernetes namespace
func (k *Kuber) ApplyConfigMap(ns string) error {
	cm := buildConfigMap(ns)
	cmInterface := k.client.CoreV1().ConfigMaps(ns)

	cm, err := cmInterface.Create(context.TODO(), cm, metav1.CreateOptions{})
	if err == nil {
		k.logger.Info("configMap created", zap.String("namespace:", ns), zap.Any("configMap:", cm))
		return nil
	} else if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	cmExists, err := cmInterface.Get(context.TODO(), cm.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cmExists.Data = cm.Data
	cm, err = cmInterface.Update(context.TODO(), cmExists, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	k.logger.Info("configMap updated", zap.String("namespace:", ns), zap.Any("configMap:", cm))
	return nil
}
