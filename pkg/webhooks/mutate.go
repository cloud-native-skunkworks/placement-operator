package webhooks

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=Ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mutate.cnskunkworks.io,admissionReviewVersions=v1,sideEffects=NoneOnDryRun
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=*;

type PodMutator struct {
	decoder *admission.Decoder
	Log     logr.Logger
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

	m.Log.Info("Admitted pod", "name", pod.Name, "namespace", pod.Namespace)

	return admission.Allowed("")
}
