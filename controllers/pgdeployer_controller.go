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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PgDeployerReconciler reconciles a PgDeployer object
type PgDeployerReconciler struct {
	Logger zlog.Logger
	client.Client
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
	logger := r.Logger.With(zap.String("reconciler", "pgDeployer"), zap.Namespace(req.Namespace))

	var pgDeploy v1alpha1.PgDeployer
	if err := r.Get(ctx, req.NamespacedName, &pgDeploy); err != nil {
		if errors.IsNotFound(err) {
			logger.Error("pg deploy not found")
			return ctrl.Result{}, err
		}
		logger.Error("could not fetch pg deploy")
		return ctrl.Result{}, err
	}

	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		logger.Error("unable to fetch Deployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := ctrl.SetControllerReference(&deployment, &deployment.ObjectMeta, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgdeployerv1alpha1.PgDeployer{}).
		Complete(r)
}
