/*
Copyright 2025.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	webappv1alpha1 "vnptplatform.vn/webapp-example/api/v1alpha1"
)

const webAppFinalizer = "webapp.vnptplatform.vn/finalizer"

// NginxWebAppReconciler reconciles a NginxWebApp object
type NginxWebAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.vnptplatform.vn,resources=nginxwebapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.vnptplatform.vn,resources=nginxwebapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.vnptplatform.vn,resources=nginxwebapps/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxWebApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *NginxWebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)

	var webapp webappv1alpha1.NginxWebApp

	if err := r.Get(ctx, req.NamespacedName, &webapp); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// ensure finalizer (optional)
	if webapp.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&webapp, webAppFinalizer) {
			controllerutil.AddFinalizer(&webapp, webAppFinalizer)
			if err := r.Update(ctx, &webapp); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// handle deletion if you need cleanup
		if controllerutil.ContainsFinalizer(&webapp, webAppFinalizer) {
			// cleanup resources if necessary
			controllerutil.RemoveFinalizer(&webapp, webAppFinalizer)
			if err := r.Update(ctx, &webapp); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Desired Deployment
	depl := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name + "-deployment",
			Namespace: webapp.Namespace,
		},
	}
	// CreateOrUpdate pattern
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, &depl, func() error {
		replicas := int32(1)
		if webapp.Spec.Replicas != nil {
			replicas = *webapp.Spec.Replicas
		}
		depl.Spec.Replicas = &replicas
		depl.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{"app": webapp.Name}}
		depl.Spec.Template.ObjectMeta.Labels = map[string]string{"app": webapp.Name}
		depl.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "nginx",
				Image: webapp.Spec.Image,
				Ports: []corev1.ContainerPort{{ContainerPort: webapp.Spec.Port}},
			},
		}
		// set owner reference
		if err := controllerutil.SetControllerReference(&webapp, &depl, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "failed to create/update Deployment")
		return ctrl.Result{}, err
	}
	logger.Info("Deployment reconciled", "operation", op)
	// Ensure Service
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name + "-svc",
			Namespace: webapp.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		svc.Spec.Selector = map[string]string{"app": webapp.Name}
		svc.Spec.Ports = []corev1.ServicePort{{Port: webapp.Spec.Port, TargetPort: intstr.FromInt(int(webapp.Spec.Port))}}
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		if err := controllerutil.SetControllerReference(&webapp, &svc, r.Scheme); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logger.Error(err, "failed to create/update Service")
		return ctrl.Result{}, err
	}

	// update status with available replicas
	var currentDep appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: webapp.Namespace, Name: depl.Name}, &currentDep); err == nil {
		logger.Info("update status")
		webapp.Status = webappv1alpha1.NginxWebAppStatus{}
		webapp.Status.AvailableReplicas = currentDep.Status.AvailableReplicas
		if currentDep.Status.AvailableReplicas < *depl.Spec.Replicas {
			webapp.Status.Phase = "Creating"
		} else {
			webapp.Status.Phase = "Running"
		}
		if err := r.Status().Update(ctx, &webapp); err != nil {
			logger.Error(err, "failed to update WebApp status")
			webapp.Status.Phase = "Error"
			// don't return error to avoid hot loops; requeue later
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxWebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1alpha1.NginxWebApp{}).
		Named("nginxwebapp").
		Complete(r)
}
