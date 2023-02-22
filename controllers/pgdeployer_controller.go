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
	"time"

	"github.com/nadavbm/pgdeployer/api/v1alpha1"
	pgdeployerv1alpha1 "github.com/nadavbm/pgdeployer/api/v1alpha1"
	"github.com/nadavbm/zlog"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	r.Logger.Info("Start reconcile", zap.String("namespace", req.NamespacedName.Namespace))

	var pgDeploy v1alpha1.PgDeployer
	if err := r.Get(ctx, req.NamespacedName, &pgDeploy); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Info("pgdeploy not found, probably deleted. skipping..", zap.String("namespace", req.Namespace))
			return ctrl.Result{Requeue: false, RequeueAfter: 0}, nil
		}
		r.Logger.Error("could not fetch v1alpha1.PgDeployer")
		return ctrl.Result{Requeue: true, RequeueAfter: time.Minute}, err
	}

	createOrUpdate := func(obj metav1.Object, object client.Object) error {
		if err := r.Get(ctx, req.NamespacedName, obj.(client.Object)); err != nil {
			if errors.IsNotFound(err) {
				r.Logger.Info("create object", zap.String("namespace", req.Namespace), zap.String("object", object.GetName()))
				if err := r.Create(ctx, object); err != nil && !errors.IsAlreadyExists(err) {
					r.Logger.Error("could not create object", zap.String("object", object.GetName()), zap.Error(err))
					return err
				}
				return nil
			}
			r.Logger.Error("unable to get", zap.String("object", object.GetName()))
			return err

		}
		return nil
	}

	var cm v1.ConfigMap
	configMap := buildConfigMap(req.Namespace, &pgDeploy)
	if err := createOrUpdate(metav1.Object(&cm), configMap); err != nil {
		return ctrl.Result{}, err
	}

	var sec v1.Secret
	secret := buildSecret(req.Namespace, &pgDeploy)
	if err := createOrUpdate(metav1.Object(&sec), secret); err != nil {
		return ctrl.Result{}, err
	}

	var dep appsv1.Deployment
	deploy := buildDeployment(req.Namespace, &pgDeploy)
	if err := createOrUpdate(metav1.Object(&dep), deploy); err != nil {
		return ctrl.Result{}, err
	}

	var svc v1.Service
	service := buildService(req.Namespace, &pgDeploy)
	if err := createOrUpdate(metav1.Object(&svc), service); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PgDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("pgdeploy-contoller").
		For(&pgdeployerv1alpha1.PgDeployer{}).
		Owns(&v1.Secret{}).
		Owns(&v1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Complete(r)
}
