package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
)

type VPNInstance struct {
	Network   *VPNNetwork
	Container *VPNContainer
}

func NewVPNInstance(name, host, user, pass string) (*VPNInstance, error) {
	net, err := NewVPNNetwork(name)
	if err != nil {
		return nil, err
	}

	ctr, err := NewVPNContainer(name, host, user, pass)
	if err != nil {
		// TODO: handle returned error?
		net.Close()
		return nil, err
	}

	return &VPNInstance{
		Network:   net,
		Container: ctr,
	}, nil
}

func (v *VPNInstance) Close() error {
	var result error

	// TODO: do i need to check err or can i just Append nil errors?
	if err := v.Container.Close(); err != nil {
		multierror.Append(result, err)
	}

	if err := v.Network.Close(); err != nil {
		multierror.Append(result, err)
	}

	return result
}

// TODO: refactor. this is ugly
func RecoverVPNInstances() (map[uuid.UUID]*VPNInstance, error) {
	networks := make(map[uuid.UUID]*VPNNetwork)
	containers := make(map[uuid.UUID]*VPNContainer)

	imageList, err := exec.Command("docker", "ps", "-q").Output()
	if err != nil {
		return nil, err
	}

	for _, ctrID := range strings.Split(string(imageList), "\n") {
		if ctrID == "" {
			continue
		}

		output, err := exec.Command("docker", "inspect", ctrID).Output()
		if err != nil {
			return nil, err
		}

		inspect := []ContainerInspectOutput{}
		if err := json.Unmarshal(output, &inspect); err != nil {
			return nil, err
		}

		labels := inspect[0].Config.Labels
		if labels[labelKey] == labelValue {
			name, err := uuid.Parse(labels["name"])
			if err != nil {
				return nil, err
			}

			ctr, err := NewVPNContainerFromID(ctrID)
			if err != nil {
				return nil, err
			}
			containers[name] = ctr
		}
	}

	networkList, err := exec.Command("docker", "network", "ls", "-q").Output()
	if err != nil {
		return nil, err
	}

	for _, networkID := range strings.Split(string(networkList), "\n") {
		if networkID == "" {
			continue
		}

		output, err := exec.Command("docker", "inspect", networkID).Output()
		if err != nil {
			return nil, err
		}

		inspect := []NetworkInspectOutput{}
		if err := json.Unmarshal(output, &inspect); err != nil {
			return nil, err
		}

		labels := inspect[0].Labels
		if labels[labelKey] == labelValue {
			name, err := uuid.Parse(labels["name"])
			if err != nil {
				return nil, err
			}

			network, err := NewVPNNetworkFromID(networkID)
			if err != nil {
				return nil, err
			}
			networks[name] = network
		}
	}

	if len(containers) != len(networks) {
		return nil, fmt.Errorf("container/network mismatch")
	}

	result := make(map[uuid.UUID]*VPNInstance)
	for name, network := range networks {
		ctr, ok := containers[name]
		if !ok {
			return nil, fmt.Errorf("failed to find container %v", name)
		}

		result[name] = &VPNInstance{
			Network:   network,
			Container: ctr,
		}
	}
	return result, nil
}
