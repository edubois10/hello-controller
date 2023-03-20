package controllers

import (
	"strings"

	machinev1 "github.com/openshift/api/machine/v1beta1"
	vsphere "github.com/openshift/machine-api-operator/pkg/controller/vsphere"
)

func getVSphereClusterPath(m *machinev1.Machine) (string, error) {

	// Here we want to get the VSphere ClusterPath.
	vsphereMachineProiderSpec, err := vsphere.ProviderSpecFromRawExtension(m.Spec.ProviderSpec.Value)
	if err != nil {
		return "", err
	}

	tmp := strings.Split(vsphereMachineProiderSpec.Workspace.ResourcePool, "/")
	clusterPath := strings.Join(tmp[:4], "/")

	return clusterPath, nil
}
