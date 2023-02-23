package layouts

import (
	corev1alpha1 "github.com/cloud-native-skunkworks/placement-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MapLayoutToAffinity(lg logr.Logger,
	layout *corev1alpha1.Layout, pod *corev1.Pod) *corev1.Affinity {
	lg.Info("Mapping layout to affinity")

	// Update the layout strategy
	pod.Labels["cnskunkworks.io/placement-operator-strategy"] = layout.Spec.Strategy

	affinity := &corev1.Affinity{}

	switch layout.Spec.Strategy {
	case "balanced":
		{
			lg.Info("Strategy is balanced")
			// Pod affinity across nodes
			affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: pod.Labels,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		}
	case "stacked":
		{
			lg.Info("Strategy is stacked")
			// Group pods together on the same node
			affinity.PodAffinity = &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: pod.Labels,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		}
	}

	return affinity
}
