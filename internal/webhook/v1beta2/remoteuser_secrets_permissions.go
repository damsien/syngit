package v1beta2

import (
	"context"
	"fmt"
	"net/http"

	syngit "github.com/syngit-org/syngit/pkg/api/v1beta2"
	utils "github.com/syngit-org/syngit/pkg/utils"
	authv1 "k8s.io/api/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

/*
	Handle webhook and get kubernetes user id
*/

type RemoteUserPermissionsWebhookHandler struct {
	Client  client.Client
	Decoder *admission.Decoder
}

func (ruwh *RemoteUserPermissionsWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	user := req.DeepCopy().UserInfo

	ru := &syngit.RemoteUser{}

	if err := utils.GetObjectFromWebhookRequest(*ruwh.Decoder, ru, req); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	namespace := ru.GetNamespace()
	if ru.Spec.SecretRef.Namespace != "" {
		namespace = ru.Spec.SecretRef.Namespace
	}

	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			User:   user.Username,
			Groups: user.Groups,
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "get",
				Group:     "",
				Version:   "v1",
				Resource:  "secrets",
				Name:      ru.Spec.SecretRef.Name,
			},
		},
	}
	err := ruwh.Client.Create(context.Background(), sar)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if !sar.Status.Allowed {
		return admission.Denied(fmt.Sprintf("The user %s is not allowed to get the secret: %s", user, ru.Spec.SecretRef.Name))
	}

	return admission.Allowed(fmt.Sprintf("The user %s is allowed to get the secret: %s", user, ru.Spec.SecretRef.Name))
}
