package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

const labelKey = "managed-by"
const labelValue = "vpnmux"

type NetworkInspectOutput struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
	IPAM struct {
		Config []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
	Labels map[string]string `json:"Labels"`
}

type VPNNetwork struct {
	Name    string
	ID      string
	Subnet  string
	Gateway string
}

func (v *VPNNetwork) String() string {
	return fmt.Sprintf(
		"Name=%s; ID=%s; Subnet=%s; Gateway=%s",
		v.Name, v.ID, v.Subnet, v.Gateway,
	)
}

func NewVPNNetwork(name string) (*VPNNetwork, error) {
	// TODO: use docker library instead
	err := exec.Command(
		"docker", "network", "create",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("name=%s", name),
		name,
	).Run()
	if err != nil {
		return nil, err
	}

	// TODO: cleanup if any of these steps fail
	return NewVPNNetworkFromID(name)
}

func NewVPNNetworkFromID(id string) (*VPNNetwork, error) {
	output, err := exec.Command("docker", "network", "inspect", id).Output()
	if err != nil {
		return nil, err
	}

	inspect := []NetworkInspectOutput{}
	if err = json.Unmarshal(output, &inspect); err != nil {
		return nil, err
	}

	// TODO: start VPN container running

	return &VPNNetwork{
		Name:    inspect[0].Name,
		ID:      inspect[0].ID,
		Subnet:  inspect[0].IPAM.Config[0].Subnet,
		Gateway: inspect[0].IPAM.Config[0].Gateway,
	}, nil
}

func (v *VPNNetwork) Close() error {
	return exec.Command("docker", "network", "rm", v.Name).Run()
}
