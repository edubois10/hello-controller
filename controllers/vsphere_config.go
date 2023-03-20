package controllers

import (
	"context"
	"fmt"

	machinev1 "github.com/openshift/api/machine/v1beta1"
	vsphere "github.com/openshift/machine-api-operator/pkg/controller/vsphere"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VSpherConfig struct {
	VSphereServer   string `yaml:"vsphere_server"`
	VSphereUser     string `yaml:"vsphere_user"`
	VSpherePassword string `yaml:"vsphere_password"`
}

func GetVSphereConfFromSecrets(client client.Client) (*VSpherConfig, error) {
	// Getting the secret with the cluster configuration
	secret, err := getVSphereSecret(client)
	if err != nil {
		fmt.Print("Not able to get the secret")
		return nil, err
	}

	server, err := getVSphereServer(client)
	if err != nil {
		fmt.Print("Not able to get vsphere server")
		fmt.Print("Not able to write the configuration")
		return nil, err
	}

	vspherConfig, err := getVSphereConfiguration(secret, server)
	if err != nil {
		return nil, err
	}
	return vspherConfig, nil
}

func getVSphereConfiguration(secret *corev1.Secret, server string) (*VSpherConfig, error) {
	vsphereConf := &VSpherConfig{
		VSphereServer:   server,
		VSphereUser:     string(secret.Data[server+".username"]),
		VSpherePassword: string(secret.Data[server+".password"]),
	}

	return vsphereConf, nil
}

func getVSphereSecret(c client.Client) (*corev1.Secret, error) {
	namespace := "openshift-machine-api"

	secret := &corev1.Secret{}
	err := c.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: "vsphere-cloud-credentials"}, secret)
	if err != nil {
		fmt.Print("Error while getting secret")
		return secret, err
	}

	return secret, nil
}

func getVSphereServer(c client.Client) (string, error) {
	machineList := &machinev1.MachineList{}
	err := c.List(context.TODO(), machineList)
	if err != nil {
		return "", err
	}
	// Here we want to get the VSphere server. Since machine in the cluster are in the same sever,
	//it doesn't matter which machine we are looking at.
	machine := machineList.Items[0]

	vsphereMachineProiderSpec, err := vsphere.ProviderSpecFromRawExtension(machine.Spec.ProviderSpec.Value)

	if err != nil {
		return "", err
	}

	return vsphereMachineProiderSpec.Workspace.Server, nil
}
