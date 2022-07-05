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

package controllers

import (
	"context"
	//	b64 "encoding/base64"
	//	"encoding/json"
	"fmt"

	//	"os"

	"github.com/grafana-tools/sdk"
	//	userv1 "github.com/openshift/api/user/v1"
	teamv1alpha1 "github.com/snapp-incubator/team-operator/api/v1alpha1"
	//	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	//	yaml "sigs.k8s.io/yaml"
)

const (
	userArgocdNS          = "user-argocd"
	userArgocRbacPolicyCM = "argocd-rbac-cm"
	userArgocStaticUserCM = "argocd-cm"
)

const (
	teamLabel = "snappcloud.io/team"
)

// Get Grafana URL and PassWord as a env.
var grafanaPassword = "xAR6WJKrszFBJsnlHCdoeuA2w2Q10y9E7iJ3J46l3Vpk1yigQl"
var grafanaUsername = "admin"
var grafanaURL = "https://grafana.okd4.teh-1.snappcloud.io"

//var logf = log.Log.WithName("controller_team")

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=team.snappcloud.io,resources=teams/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=user.openshift.io,resources=*,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the clus k8s.io/api closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Team object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *TeamReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling team")
	team := &teamv1alpha1.Team{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, team)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	} else {
		log.Info("team is found and teamName is : " + team.Name)

	}
	//r.createArgocdStaticAdminUser(ctx, req)
	//r.createArgocdStaticViewUser(ctx, req)
	r.AddUsersToGrafanaOrgByEmail(ctx, req, team.Spec.Grafana.Admin.Emails, "admin")
	r.AddUsersToGrafanaOrgByEmail(ctx, req, team.Spec.Grafana.Edit.Emails, "edit")
	r.AddUsersToGrafanaOrgByEmail(ctx, req, team.Spec.Grafana.View.Emails, "view")

	return ctrl.Result{}, nil
}
func (r *TeamReconciler) AddUsersToGrafanaOrgByEmail(ctx context.Context, req ctrl.Request, emails []string, role string) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling team")
	team := &teamv1alpha1.Team{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, team)
	if err != nil {
		log.Error(err, "Failed to get sina")
		return ctrl.Result{}, err
	}
	org := team.GetLabels()[teamLabel]
	fmt.Println(org)
	// Connecting to the Grafana API
	client, err1 := sdk.NewClient(grafanaURL, fmt.Sprintf("%s:%s", grafanaUsername, grafanaPassword), sdk.DefaultHTTPClient)
	if err1 != nil {
		log.Error(err1, "Unable to create Grafana client")
		return ctrl.Result{}, err1
	} else {
		for _, email := range emails {
			retrievedOrg, _ := client.GetOrgByOrgName(ctx, org)
			orgID := retrievedOrg.ID
			newuser := sdk.UserRole{LoginOrEmail: email, Role: role}
			_, err := client.AddOrgUser(ctx, newuser, orgID)
			if err != nil {
				log.Error(err, "Failed to add user to  organization")
				return ctrl.Result{}, err
			} else {
				log.Info(email, "added", "for team", team, "organization name is", retrievedOrg.Name)
				return ctrl.Result{}, nil
			}
		}
		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&teamv1alpha1.Team{}).
		Complete(r)
}
