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
	"k8s.io/apimachinery/pkg/types"
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

	var pgDeploy v1alpha1.PgDeployer
	err := r.Get(ctx, req.NamespacedName, &pgDeploy)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Error("pg deploy not found")
			return ctrl.Result{}, err
		}
		r.Logger.Error("could not fetch pg deploy")
		return ctrl.Result{}, err
	}

	cm, err := r.buildConfigMap(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build configmap")
		return ctrl.Result{}, err
	}
	r.Logger.Info("create configmap", zap.String("namespace", req.Namespace), zap.String("name", cm.Name))
	err = r.Get(ctx, types.NamespacedName{Name: pgDeploy.Name, Namespace: pgDeploy.Namespace}, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			if err = r.Create(ctx, cm); err != nil {
				r.Logger.Error("could not create a configmap")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	secret, err := r.buildSecret(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build secret")
		return ctrl.Result{}, err
	}
	r.Logger.Info("create pg-secret", zap.String("namespace", req.Namespace), zap.String("name", secret.Name))
	err = r.Get(ctx, types.NamespacedName{Name: pgDeploy.Name, Namespace: pgDeploy.Namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			if err = r.Create(ctx, secret); err != nil {
				r.Logger.Error("could not create a secret")
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	deploy, err := r.buildDeployment(req.Namespace, testing, &pgDeploy)
	if err != nil {
		r.Logger.Error("could not build deployment")
		return ctrl.Result{}, err
	}
	r.Logger.Info("create deployment", zap.String("namespace", req.Namespace), zap.String("name", deploy.Name))
	err = r.Get(ctx, types.NamespacedName{Name: pgDeploy.Name, Namespace: pgDeploy.Namespace}, deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			if err = r.Create(ctx, deploy); err != nil {
				r.Logger.Error("could not create a deployment")
				return ctrl.Result{}, err
			}
		} else {
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
