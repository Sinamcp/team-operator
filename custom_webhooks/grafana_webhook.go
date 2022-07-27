package custom_webhook

import (
	"context"
	"fmt"

	"github.com/grafana-tools/sdk"
	grafanav1alpha1 "github.com/snapp-incubator/team-operator/apis/grafana/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/validate-v1-resource-quota,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=resourcequotas,verbs=create;update;delete,versions=v1,name=vresourcequota.kb.io,admissionReviewVersions={v1,v1beta1}

type GrafanaValidator struct {
	Client client.Client
}

// Get Grafana URL and PassWord as a env.
var grafanaPassword = "xAR6WJKrszFBJsnlHCdoeuA2w2Q10y9E7iJ3J46l3Vpk1yigQl"
var grafanaUsername = "admin"
var grafanaURL = "https://grafana.okd4.teh-1.snappcloud.io"

func (v *GrafanaValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	grafana := &grafanav1alpha1.Grafana{}
	responseAdmin := v.ValidateEmailExist(ctx, req, grafana.Spec.Admin)
	responseEdit := v.ValidateEmailExist(ctx, req, grafana.Spec.Edit)
	responseView := v.ValidateEmailExist(ctx, req, grafana.Spec.View)

	if responseAdmin.AdmissionResponse.Allowed == true && responseEdit.AdmissionResponse.Allowed == true && responseView.AdmissionResponse.Allowed == true {
		return admission.Allowed("Allowed")
	} else {
		return admission.Denied("Denied")
	}
}

func (v *GrafanaValidator) ValidateEmailExist(ctx context.Context, req admission.Request, emails []string) admission.Response {
	client, _ := sdk.NewClient(grafanaURL, fmt.Sprintf("%s:%s", grafanaUsername, grafanaPassword), sdk.DefaultHTTPClient)
	grafanalUsers, _ := client.GetAllUsers(ctx)
	var Users = make([]string, 10)
	for _, email := range emails {
		for _, grafanauser := range grafanalUsers {
			Grafanau := grafanauser.Email
			if email == Grafanau {
				continue
			} else {
				Users = append(Users, Grafanau)

			}

		}
	}
	if len(Users) == 0 {
		return admission.Denied("please make sure all of the user are login")

	}
	return admission.Allowed("user will be added")
}
