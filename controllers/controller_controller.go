/*
Copyright 2023.

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
	"fmt"

	machinev1 "github.com/openshift/api/machine/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// DRSVmGroupReconciler reconciles a DRSVmGroup object
type DRSVmGroupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Server   string
	Username string
	Password string
}

// +kubebuilder:rbac:groups=machine.openshift.io,resources=machines,verbs=get;list;watch;
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DRSVmGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *DRSVmGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := ctrllog.FromContext(ctx)

	// TODO(user): your logic here

	// GET all the machine with the label vmware.bit.admin.ch/drs-vm-group
	machine := &machinev1.Machine{}
	err := r.Get(ctx, req.NamespacedName, machine)
	if err != nil {
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Machine")
		return ctrl.Result{}, err
	}

	// Checking that the machine we are reconciling has the vmware.bit.admin.ch/drs-vm-group label.
	vmGroupName, ok := machine.ObjectMeta.Labels["vmware.bit.admin.ch/drs-vm-group"]
	if ok {
		// Get Machine Name
		machineName := machine.ObjectMeta.Name
		fmt.Printf("Name: %s", machineName)
		// Get Machine Cluster Path
		clusterPath, err := getVSphereClusterPath(machine)
		fmt.Printf("Name: %s", clusterPath)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Create new vim25.Client to connect to VSphere
		client25, err := r.NewClient(ctx, true)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Check the machine is in the group
		err = addVmToGroup(ctx, client25, machineName, vmGroupName, clusterPath)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Add the VM to the v
	}

	//machines := &machinev1.MachineList{}

	// Loop over the machines with the wanted label

	// Check if the

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DRSVmGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&machinev1.Machine{}).
		Complete(r)
}
