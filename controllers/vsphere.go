package controllers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// VMGroup represents a VSphere VM Group object.
type VMGroup struct {
	Cluster *object.ClusterComputeResource
	VmGroup *types.ClusterVmGroup
}

// HasVM returns whether a VSphere VM object is a member of the VM Group.
func (vg VMGroup) HasVM(vmObj types.ManagedObjectReference) (bool, error) {
	vms := vg.listVMs()

	for _, vm := range vms {
		if vm == vmObj {
			return true, nil
		}
	}
	return false, nil
}

func (vg VMGroup) listVMs() []types.ManagedObjectReference {
	return vg.VmGroup.Vm
}

func FindVMGroup(ctx context.Context, c *vim25.Client, clusterPath string, vmGroupName string) (*VMGroup, error) {
	// Manager is an object that provides management capabilities for a specific set of vSphere objects.
	m := view.NewManager(c)
	/*
		A ContainerView is a way to filter and retrieve a subset of objects from a vSphere inventory. The function takes a single parameter,
		which is a reference to a vSphere object such as a Datacenter, and returns a new ContainerView object that contains all the child objects
		of the specified object, including virtual machines, networks, datastores, etc. The returned ContainerView object can then be used to
		filter and retrieve specific objects based on their type and properties.
	*/
	_, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		return nil, err
	}

	// Create Finder
	finder := find.NewFinder(c, false)

	// Find the ClusterComputeResource by its path
	clusterComputeResourceList, err := finder.ClusterComputeResourceList(ctx, clusterPath) // The path can be retrieved from the machine crd
	if err != nil {
		fmt.Printf("Error occurred while getting clusterComputeResource %+v", err.Error())
		return nil, err
	}

	cluster := clusterComputeResourceList[0]
	clusterConfigInfoEx, err := cluster.Configuration(ctx)
	if err != nil {
		return nil, err
	}

	for _, group := range clusterConfigInfoEx.Group {
		if clusterVMGroup, ok := group.(*types.ClusterVmGroup); ok {
			if clusterVMGroup.Name == vmGroupName {
				return &VMGroup{cluster, clusterVMGroup}, nil
			}
		}
	}
	return nil, errors.Errorf("cannot find VM group %s", vmGroupName)
}

func FindVm(ctx context.Context, c *vim25.Client, vmName string) (*types.ManagedObjectReference, error) {

	var vm *types.ManagedObjectReference
	// Manager is an object that provides management capabilities for a specific set of vSphere objects
	m := view.NewManager(c)
	/*
		A ContainerView is a way to filter and retrieve a subset of objects from a vSphere inventory. The function takes a single parameter,
		which is a reference to a vSphere object such as a Datacenter, and returns a new ContainerView object that contains all the child objects
		of the specified object, including virtual machines, networks, datastores, etc. The returned ContainerView object can then be used to
		filter and retrieve specific objects based on their type and properties.
	*/
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return vm, err
	}

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return vm, err
	}

	// Print summary per vm (see also: govc/vm/info.go)

	for _, vm := range vms {
		if vm.Summary.Config.Name == vmName {
			fmt.Printf("%s\n", vm.Summary.Config.Name)
			fmt.Printf("Type of Vm is %T\n", vm)
			return vm.Summary.Vm, nil
		}
	}

	return vm, errors.Errorf("cannot find VM with name %s", vmName)
}

// Add a VSphere VM object to the VM Group.
func (vg VMGroup) Add(ctx context.Context, vmObj types.ManagedObjectReference) (*object.Task, error) {
	vms := vg.listVMs()
	vg.VmGroup.Vm = append(vms, vmObj) //nolint:gocritic

	spec := &types.ClusterConfigSpecEx{
		GroupSpec: []types.ClusterGroupSpec{
			{
				ArrayUpdateSpec: types.ArrayUpdateSpec{
					Operation: types.ArrayUpdateOperationEdit,
				},
				Info: vg.VmGroup,
			},
		},
	}
	return vg.Cluster.Reconfigure(ctx, spec, true)
}

func addVmToGroup(ctx context.Context, c *vim25.Client, machineName string, vmGroup string, clusterPath string) error {

	// Getting the vmGroup object by name
	vmGroupObj, err := FindVMGroup(ctx, c, clusterPath, vmGroup)
	if err != nil {
		return err
	}
	// TMP
	vmList := vmGroupObj.listVMs()
	fmt.Printf("VmList: %s", vmList)

	// Getting the vm object by it's name
	vm_, err := FindVm(ctx, c, machineName)
	if err != nil {
		return err
	}
	vm := *vm_

	// Checking if the vm already belongs to the vmGroup
	vmInGroup, err := vmGroupObj.HasVM(vm)
	if err != nil {
		return err
	}
	// The vm is already part of the VmGroup so no need to add it again
	if vmInGroup {
		return nil
	}
	// Adding the Vm to the VmGroup
	vmGroupObj.Add(ctx, vm)
	return nil
}

func processOverride(u *url.URL, username string, password string) {
	u.User = url.User(username)
	u.User = url.UserPassword(username, password)
}

// NewClient creates a vim25.Client for use in the examples
func (r *DRSVmGroupReconciler) NewClient(ctx context.Context, insecure bool) (*vim25.Client, error) {

	//url := &r.Server
	// Parse URL from string
	u, err := soap.ParseURL(r.Server) // probably not needed, it's to get the user and password from the url
	if err != nil {
		return nil, err
	}

	// Override username and/or password as required
	processOverride(u, r.Username, r.Password)

	// Share govc's session cache
	s := &cache.Session{
		URL:         u,
		Insecure:    insecure,
		Passthrough: true, // Passthrough disables caching when set to true
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		errors.Errorf("Failed to Loggin with cached data: %s", err)
		return nil, err
	}

	return c, nil
}
