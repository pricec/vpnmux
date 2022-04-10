package network

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/openvpn"
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

type Network struct {
	ID        string
	DockerID  string
	Subnet    string
	Gateway   string
	Container *Container
}

func New(id string, cfg *openvpn.Config) (*Network, error) {
	// TODO: use docker library instead
	err := exec.Command(
		"docker", "network", "create",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("id=%s", id),
		id,
	).Run()
	if err != nil {
		return nil, err
	}

	_, err = NewContainer(id, cfg)
	if err != nil {
		return nil, err
	}

	// TODO: cleanup if any of these steps fail
	return NewFromID(id)
}

func NewFromID(id string) (*Network, error) {
	ctr, err := NewContainerFromID(id)
	if err != nil {
		return nil, err
	}

	output, err := exec.Command("docker", "network", "ls", "-q", "--filter", fmt.Sprintf("label=id=%s", id)).Output()
	if err != nil {
		return nil, err
	}

	dockerID := string(output[:len(output)-1])
	output, err = exec.Command("docker", "network", "inspect", dockerID).Output()
	if err != nil {
		return nil, err
	}

	inspect := []NetworkInspectOutput{}
	if err = json.Unmarshal(output, &inspect); err != nil {
		return nil, err
	}

	return &Network{
		ID:        id,
		DockerID:  inspect[0].ID,
		Subnet:    inspect[0].IPAM.Config[0].Subnet,
		Gateway:   inspect[0].IPAM.Config[0].Gateway,
		Container: ctr,
	}, nil
}

func (v *Network) Close() error {
	var result error

	if v.Container != nil {
		if err := v.Container.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	if err := exec.Command("docker", "network", "rm", v.DockerID).Run(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
