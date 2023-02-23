/*
Copyright 2023 AlexsJones.

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

	corev1alpha1 "github.com/cloud-native-skunkworks/placement-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LayoutReconciler reconciles a Layout object
type LayoutReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Layouts map[string]*corev1alpha1.Layout
}

//+kubebuilder:rbac:groups=core.cnskunkworks.io,resources=layouts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.cnskunkworks.io,resources=layouts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.cnskunkworks.io,resources=layouts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Layout object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *LayoutReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)

	// get layout
	layout := &corev1alpha1.Layout{}
	if err := r.Get(ctx, req.NamespacedName, layout); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Check if layout is in map and add if not
	if _, ok := r.Layouts[layout.Name]; !ok {
		r.Layouts[layout.Name] = layout
		lg.Info("Added layout to map", "layout", layout.Name)
	} else {
		if r.Layouts[layout.Name].Spec.Strategy != layout.Spec.Strategy {
			r.Layouts[layout.Name].Spec.Strategy = layout.Spec.Strategy
			lg.Info("Updated layout in map", "layout", layout.Name)

			// List pods that are using annotation cnskunkworks.io/placement-operator-layout
			pods := &corev1.PodList{}
			if err := r.List(ctx, pods, client.MatchingLabels{"cnskunkworks.io/placement-operator-layout": layout.Name}); err != nil {
				lg.Error(err, "Failed to list pods")
				return ctrl.Result{}, err
			}
			lg.Info("Found pods", "pods", pods.Items)
			// Probably need to find their owners and restart
			// Iterate through pods and update their node affinity
			for _, pod := range pods.Items {
				// Remove the pod
				if err := r.Delete(ctx, &pod); err != nil {
					lg.Error(err, "Failed to delete pod", "pod", pod.Name)
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LayoutReconciler) SetupWithManager(mgr ctrl.Manager) error {

	r.Layouts = make(map[string]*corev1alpha1.Layout, 0)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Layout{}).
		Complete(r)
}
