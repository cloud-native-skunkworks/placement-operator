package webhooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cloud-native-skunkworks/placement-operator/controllers"
	"github.com/cloud-native-skunkworks/placement-operator/pkg/webhooks/layouts"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.cnskunkworks.io,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*;

type PodMutator struct {
	client.Client
	decoder             *admission.Decoder
	Log                 logr.Logger
	LayoutReconcilerRef *controllers.LayoutReconciler
}

func (m *PodMutator) InjectDecoder(d *admission.Decoder) error {
	m.decoder = d
	return nil
}

func (m *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	m.Log.Info("Handling pod", "name", req.Name, "namespace", req.Namespace)
	pod := &corev1.Pod{}
	err := m.decoder.Decode(req, pod)
	if err != nil {
		m.Log.Error(err, "unable to decode pod")
		return admission.Errored(http.StatusBadRequest, err)
	}
	// Check if the pod is an orphan
	if pod.OwnerReferences == nil {
		m.Log.Info("Ignoring pod", "name", pod.Name, "namespace", pod.Namespace, "reason", "orphan")
		return admission.Allowed("")
	}
	// Check if the placement-operator is enabled for the workload
	if pod.Labels["cnskunkworks.io/placement-operator-enabled"] == "true" {
		m.Log.Info("Placement enabled for pod", "name", pod.Name, "namespace", pod.Namespace)
		// Get the placement strategy and apply to the pod
		layoutName := "default"
		if pod.Labels["cnskunkworks.io/placement-operator-layout"] != "" {
			layoutName = pod.Labels["cnskunkworks.io/placement-operator-layout"]
		}
		// Check map is not null
		if m.LayoutReconcilerRef.Layouts != nil {
			// Check if the layout exists
			if layout, ok := m.LayoutReconcilerRef.Layouts[layoutName]; ok {
				// Set pod affinity based on layout
				pod.Spec.Affinity = layouts.MapLayoutToAffinity(m.Log, layout, pod)
				// Implement changes
				b, err := json.Marshal(pod)
				if err != nil {
					m.Log.Error(err, "unable to marshal pod")
					return admission.Errored(http.StatusBadRequest, err)
				}
				return admission.PatchResponseFromRaw(req.Object.Raw, b)

			} else {
				m.Log.Error(err, "Layout not found", "layout", layoutName)
			}
		} else {
			m.Log.Error(err, "Layouts map is null")
		}

	} else {
		m.Log.Info("Placement disabled for pod", "name", pod.Name, "namespace", pod.Namespace)
	}
	return admission.Allowed("")
}

// func (m *PodMutator) isWorkloadEnabledForPlacementOperator(pod *corev1.Pod) (bool, error) {

// 	// Get the owner of the pod
// 	m.Log.Info("Checking owner ref")
// 	owner := pod.OwnerReferences[0]
// 	// Check if replica set
// 	if owner.Kind == "ReplicaSet" {
// 		m.Log.Info("Owner ref is a replica set, finding owner above")
// 		// Get the replica set
// 		rs := &appsv1.ReplicaSet{}
// 		if err := m.Client.Get(context.Background(),
// 			types.NamespacedName{Name: owner.Name, Namespace: pod.Namespace}, rs); err != nil {
// 			return false, err
// 		}
// 		// Get the owner of the replica set
// 		owner = rs.OwnerReferences[0]
// 	}
// 	m.Log.Info("Owner ref", "kind", owner.Kind, "name", owner.Name, "uid", owner.UID)
// 	// Get owner by type
// 	switch owner.Kind {
// 	case "Deployment":
// 		// Get the deployment
// 		deployment := &appsv1.Deployment{}
// 		if err := m.Client.Get(context.Background(),
// 			types.NamespacedName{Name: owner.Name, Namespace: pod.Namespace}, deployment); err != nil {
// 			return false, err
// 		}
// 		// Check if the deployment has the placement annotation
// 		if _, ok := deployment.Annotations["cnskunkworks.io/placement"]; ok {
// 			return true, nil
// 		}
// 	case "StatefulSet":
// 		// Get the stateful set
// 		statefulSet := &appsv1.StatefulSet{}
// 		if err := m.Client.Get(context.Background(),
// 			types.NamespacedName{Name: owner.Name, Namespace: pod.Namespace}, statefulSet); err != nil {
// 			return false, err
// 		}
// 		// Check if the stateful set has the placement annotation
// 		if _, ok := statefulSet.Annotations["cnskunkworks.io/placement"]; ok {
// 			return true, nil
// 		}
// 	}

// 	return false, nil
// }
