/*
Copyright 2022.

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

package grafana

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana-tools/sdk"
	grafanav1alpha1 "github.com/snapp-incubator/team-operator/apis/grafana/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	teamLabel = "snappcloud.io/team"
)

// Get Grafana URL and PassWord as a env.
var grafanaPassword = "xAR6WJKrszFBJsnlHCdoeuA2w2Q10y9E7iJ3J46l3Vpk1yigQl"
var grafanaUsername = "admin"
var grafanaURL = "https://grafana.okd4.teh-1.snappcloud.io"

// GrafanaReconciler reconciles a Grafana object
type GrafanaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.snappcloud.io,resources=grafanas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.snappcloud.io,resources=grafanas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.snappcloud.io,resources=grafanas/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=user.openshift.io,resources=*,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Grafana object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *GrafanaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)

	reqLogger.Info("Reconciling grafana11")
	grafana := &grafanav1alpha1.Grafana{}
	r.AddUsersToGrafanaOrgByEmail(ctx, req, grafana.Spec.Admin.Emails, "admin")
	r.AddUsersToGrafanaOrgByEmail(ctx, req, grafana.Spec.Edit.Emails, "editor")
	r.AddUsersToGrafanaOrgByEmail(ctx, req, grafana.Spec.View.Emails, "viewer")

	return ctrl.Result{}, nil
}
func (r *GrafanaReconciler) AddUsersToGrafanaOrgByEmail(ctx context.Context, req ctrl.Request, emails []string, role string) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	ns := &corev1.Namespace{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: req.Namespace}, ns)

	if err != nil {
		log.Error(err, "Failed to get namespace")
		return ctrl.Result{}, err
	}
	org := ns.GetLabels()[teamLabel]
	// Connecting to the Grafana API
	client, err1 := sdk.NewClient(grafanaURL, fmt.Sprintf("%s:%s", grafanaUsername, grafanaPassword), sdk.DefaultHTTPClient)
	retrievedOrg, _ := client.GetOrgByOrgName(ctx, org)
	orgID := retrievedOrg.ID
	getallUser, _ := client.GetAllUsers(ctx)
	getuserOrg, _ := client.GetOrgUsers(ctx, orgID)
	if err1 != nil {
		log.Error(err1, "Unable to create Grafana client")
		return ctrl.Result{}, err1
	} else {
		for _, email := range emails {
			reqLogger.Info("1")
			var orguserfound bool
			for _, orguser := range getuserOrg {
				UserOrg := orguser.Email
				if email == UserOrg {
					reqLogger.Info("2")
					orguserfound = true
					reqLogger.Info("users already in")
					break
				}
			}
			if orguserfound {
				continue
			}
			for _, user := range getallUser {
				reqLogger.Info("3")
				UserEmail := user.Email
				if email == UserEmail {
					reqLogger.Info("user is exist")
					reqLogger.Info("4")
					newuser := sdk.UserRole{LoginOrEmail: email, Role: role}
					_, err := client.AddOrgUser(ctx, newuser, orgID)
					if err != nil {
						log.Error(err, "Failed to add user to  organization")
					} else {
						log.Info("ok")
					}
					break
				}
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1alpha1.Grafana{}).
		Complete(r)
}
