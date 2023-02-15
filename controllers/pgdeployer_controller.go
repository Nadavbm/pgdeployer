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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PgDeployerReconciler reconciles a PgDeployer object
type PgDeployerReconciler struct {
	Logger *zlog.Logger
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
	logger := zlog.New()
	r.Logger = logger

	r.Logger.Info("v1alpha1.PgDeployer changed. Start reconcile", zap.String("namespace", req.NamespacedName.Namespace))

	var pgDeploy v1alpha1.PgDeployer
	if err := r.Get(ctx, req.NamespacedName, &pgDeploy); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Info("pg deploy not found, probably deleted. skipping..", zap.String("namespace", req.Namespace))
			return ctrl.Result{}, nil
		}
		r.Logger.Error("could not fetch v1alpha1.PgDeployer, check if crd applied in the cluster..")
		return ctrl.Result{}, err
	}

	objects := pgDeploy.Construct(req.Namespace)

	for _, object := range objects {
		if err := controllerutil.SetControllerReference(&pgDeploy, object, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		r.Logger.Info("create object", zap.String("namespace", req.Namespace), zap.String("object", object.GetName()))
		if err := r.Create(ctx, object.(client.Object)); err != nil {
			if errors.IsAlreadyExists(err) {
				if err := r.Update(ctx, object.(client.Object)); err != nil {
					if errors.IsInvalid(err) {
						r.Logger.Error("invalid update", zap.String("object", object.GetName()))
					} else {
						r.Logger.Error("unable to update", zap.String("object", object.GetName()))
					}
				}
			} else {
				r.Logger.Error("could not create object", zap.String("object", object.GetName()), zap.Error(err))
				return ctrl.Result{}, err
			}
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
