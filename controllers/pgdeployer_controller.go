/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"
	pgdeployerv1alpha1 "github.com/nadavbm/pgdeployer/api/v1alpha1"
	"github.com/nadavbm/zlog"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PgDeployerReconciler reconciles a PgDeployer object
type PgDeployerReconciler struct {
	Logger *zlog.Logger
	client.Client
	client kubernetes.Clientset
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pgdeployer.example.com,resources=pgdeployers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pgdeployer.example.com,resources=pgdeployers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pgdeployer.example.com,resources=pgdeployers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PgDeployer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PgDeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	testing := false
	logger := zlog.New()
	r.Logger = logger

	r.Logger.Info("v1alpha1.PgDeployer added. Start reconcile", zap.String("namespace", req.Namespace))

	var pgDeploy v1alpha1.PgDeployer
	if err := r.Get(ctx, req.NamespacedName, &pgDeploy); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Info("pg deploy not found, probably deleted. skipping..", zap.String("namespace", req.Namespace))
			return ctrl.Result{}, nil
		}
		r.Logger.Error("could not fetch v1alpha1.PgDeployer, check if crd applied in the cluster..")
		return ctrl.Result{}, err
	}

	r.Logger.Info("create configmap", zap.String("namespace", req.Namespace))
	// config map creation
	cm, err := r.buildConfigMap(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not create configmap specs", zap.Error(err))
		return ctrl.Result{}, err
	}

	if err = r.Create(ctx, cm); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, cm); err != nil {
				if errors.IsInvalid(err) {
					r.Logger.Error("invalid config map update")
				} else {
					r.Logger.Error("unable to update configmap")
				}
			}
		} else {
			r.Logger.Error("could not create a configmap", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	r.Logger.Info("create pg-secret", zap.String("namespace", req.Namespace))
	// secret creation
	secret, err := r.buildSecret(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build secret", zap.Error(err))
		return ctrl.Result{}, err
	}

	if err = r.Create(ctx, secret); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, secret); err != nil {
				if errors.IsInvalid(err) {
					r.Logger.Error("invalid secret update")
				} else {
					r.Logger.Error("unable to update secret")
				}
			}
		} else {
			r.Logger.Error("could not create secret", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	r.Logger.Info("create deployment", zap.String("namespace", req.Namespace))
	// postgres deployment creation
	deploy, err := r.buildDeployment(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build deployment", zap.Error(err))
		return ctrl.Result{}, err
	}

	if err = r.Create(ctx, deploy); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, secret); err != nil {
				if errors.IsInvalid(err) {
					r.Logger.Error("invalid deployment update")
				} else {
					r.Logger.Error("unable to update deployment")
				}
			}
		} else {
			r.Logger.Error("could not create a deployment", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	r.Logger.Info("create service", zap.String("namespace", req.Namespace))
	// service creation
	svc, err := r.buildService(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build service", zap.Error(err))
		return ctrl.Result{}, err
	}

	if err = r.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, secret); err != nil {
				if errors.IsInvalid(err) {
					r.Logger.Error("invalid service update")
				} else {
					r.Logger.Error("unable to update service")
				}
			}
		} else {
			r.Logger.Error("could not create service", zap.Error(err))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgdeployerv1alpha1.PgDeployer{}).
		Complete(r)
}
