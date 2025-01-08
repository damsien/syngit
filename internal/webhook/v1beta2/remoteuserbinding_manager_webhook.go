package v1beta2

import (
	"context"
	"fmt"
	"net/http"

	syngit "github.com/syngit-org/syngit/pkg/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

/*
	Handle webhook and get kubernetes user id
*/

type RemoteUserBindingManagerWebhookHandler struct {
	Client         client.Client
	Decoder        *admission.Decoder
	Namespace      string
	ServiceAccount string
	syngitSa       string
	username       string
}

const (
	serviceAccountPrefix = "system:serviceaccount:"
)

// +kubebuilder:webhook:path=/syngit-v1beta2-remoteuserbinding-manager,mutating=false,failurePolicy=fail,sideEffects=None,groups=syngit.io,resources=remoteuserbindings,verbs=create;update;delete,versions=v1beta2,admissionReviewVersions=v1,name=vremoteuserbindings-manager.v1beta2.syngit.io

func (rubwh *RemoteUserBindingManagerWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	_ = log.FromContext(ctx)

	rubwh.syngitSa = fmt.Sprintf("%s%s:%s", serviceAccountPrefix, rubwh.Namespace, rubwh.ServiceAccount)
	rubwh.username = req.DeepCopy().UserInfo.Username
	log.Log.Info("INFOOO")
	log.Log.Info(rubwh.syngitSa)
	log.Log.Info(rubwh.username)
	var labels map[string]string
	rub := &syngit.RemoteUserBinding{}

	switch string(req.Operation) {
	case "DELETE":
		{
			err := rubwh.Decoder.DecodeRaw(req.OldObject, rub)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			labels = rub.Labels

			if rubwh.labelChecker(labels) {
				return admission.Denied("The object is managed by Syngit. It cannot be deleted by another user than the Syngit ServiceAccount.")
			}
		}
	case "UPDATE":
		{
			err := rubwh.Decoder.DecodeRaw(req.OldObject, rub)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			labels = rub.Labels
			if rubwh.labelChecker(labels) {
				return admission.Denied("The object is managed by Syngit. It cannot be deleted by another user than the Syngit ServiceAccount.")
			}

			err = rubwh.Decoder.Decode(req, rub)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			labels = rub.Labels
			if rubwh.labelChecker(labels) {
				return admission.Denied("The object is not supposed to be managed by Syngit. Only the Syngit ServiceAccount can set the resource to be managed by Syngit.")
			}
		}
	case "CREATE":
		{
			err := rubwh.Decoder.Decode(req, rub)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			labels = rub.Labels
			if rubwh.labelChecker(labels) {
				return admission.Denied("The object is not supposed to be managed by Syngit. Only the Syngit ServiceAccount can set the resource to be managed by Syngit. " + rubwh.username)
			}
		}
	}

	return admission.Allowed("This object is not managed by Syngit.")
}

func (rubwh *RemoteUserBindingManagerWebhookHandler) labelChecker(labels map[string]string) bool {
	if labels[managedByLabelKey] == managedByLabelValue {
		if rubwh.username != rubwh.syngitSa {
			return true
		}
	}
	return false
}
