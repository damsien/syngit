package v1beta3

import (
	"context"
	"net/http"
	"time"

	patterns "github.com/syngit-org/syngit/internal/patterns/v1beta3"
	syngit "github.com/syngit-org/syngit/pkg/api/v1beta3"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

/*
	Handle webhook and get kubernetes user id
*/

type RemoteUserAssociationWebhookHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
}

// +kubebuilder:webhook:path=/syngit-v1beta3-remoteuser-association,mutating=false,failurePolicy=fail,sideEffects=None,groups=syngit.io,resources=remoteusers,verbs=create;update;delete,versions=v1beta3,admissionReviewVersions=v1,name=vremoteusers-association.v1beta3.syngit.io

func (ruwh *RemoteUserAssociationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	var remoteUser *syngit.RemoteUser
	var isEnabled = false

	if string(req.Operation) == "DELETE" { //nolint:goconst
		remoteUser = &syngit.RemoteUser{}
		err := ruwh.Decoder.DecodeRaw(req.OldObject, remoteUser)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
	} else {
		remoteUser = &syngit.RemoteUser{}
		err := ruwh.Decoder.Decode(req, remoteUser)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if remoteUser.Annotations[syngit.RubAnnotation] == "true" {
			isEnabled = true
		}
	}

	username := req.DeepCopy().UserInfo.Username
	associationPattern := &patterns.RemoteUserAssociationPattern{
		PatternSpecification: patterns.PatternSpecification{
			Client:         ruwh.Client,
			NamespacedName: types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		},
		Username:   username,
		RemoteUser: *remoteUser,
		IsEnabled:  isEnabled,
	}
	remoteTargetPattern := &patterns.RemoteUserSearchRemoteTargetPattern{
		PatternSpecification: patterns.PatternSpecification{
			Client:         ruwh.Client,
			NamespacedName: types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		},
		Username:   username,
		RemoteUser: *remoteUser,
		IsEnabled:  isEnabled,
	}

	err := patterns.Trigger(associationPattern, ctx)
	if err != nil {
		if err.Reason == patterns.Denied {
			return admission.Denied(err.Message)
		}
		if err.Reason == patterns.Errored {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	// TODO: temporary workaround
	// We need to wait until the associated RemoteUserBinding
	// is fully created to allow the association with the managed RemoteTargets
	time.Sleep(1000 * time.Millisecond)

	err = patterns.Trigger(remoteTargetPattern, ctx)
	if err != nil {
		if err.Reason == patterns.Denied {
			return admission.Denied(err.Message)
		}
		if err.Reason == patterns.Errored {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	userSpecificError := ruwh.triggerUserSpecificPatterns(ctx, req, username)
	if userSpecificError != nil {
		if userSpecificError.Reason == patterns.Denied {
			return admission.Denied(userSpecificError.Message)
		}
		if userSpecificError.Reason == patterns.Errored {
			return admission.Errored(http.StatusInternalServerError, userSpecificError)
		}
	}

	return admission.Allowed("This object is associated to the " + req.Name + " RemoteUserBinding")
}

func (ruwh *RemoteUserAssociationWebhookHandler) triggerUserSpecificPatterns(ctx context.Context, req admission.Request, username string) *patterns.ErrorPattern {
	// Get all RemoteSyncer of the namespace that implement the user specific pattern
	remoteSyncerList := &syngit.RemoteSyncerList{}
	selector := labels.NewSelector()
	userSpecificKey, reqErr := labels.NewRequirement(syngit.RtAnnotationUserSpecificKey, selection.Exists, nil)
	if reqErr != nil {
		return &patterns.ErrorPattern{Message: reqErr.Error(), Reason: patterns.Errored}
	}
	managedBy, reqErr := labels.NewRequirement(syngit.RtAnnotationUserSpecificKey, selection.Equals, []string{syngit.ManagedByLabelValue})
	if reqErr != nil {
		return &patterns.ErrorPattern{Message: reqErr.Error(), Reason: patterns.Errored}
	}
	selector.Add(*userSpecificKey)
	selector.Add(*managedBy)
	listOps := &client.ListOptions{
		LabelSelector: selector,
		Namespace:     req.Namespace,
	}
	listErr := ruwh.Client.List(ctx, remoteSyncerList, listOps)
	if listErr != nil {
		return &patterns.ErrorPattern{Message: listErr.Error(), Reason: patterns.Errored}
	}

	for _, rsy := range remoteSyncerList.Items {
		userSpecificPattern := &patterns.UserSpecificPattern{
			PatternSpecification: patterns.PatternSpecification{
				Client:         ruwh.Client,
				NamespacedName: types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
			},
			Username:     username,
			RemoteSyncer: rsy,
		}

		err := patterns.Trigger(userSpecificPattern, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
