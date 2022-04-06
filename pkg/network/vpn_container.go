package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

const imageName = "openvpn-client"

type ContainerInspectOutput struct {
	ID    string `json:"Id"`
	State struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		Paused     bool   `json:"Paused"`
		Restarting bool   `json:"Restarting"`
		OOMKilled  bool   `json:"OOMKilled"`
		Dead       bool   `json:"Dead"`
		Pid        int    `json:"Pid"`
		ExitCode   int    `json:"ExitCode"`
		Error      string `json:"Error"`
	} `json:"State"`
	Image  string `json:"Image"`
	Name   string `json:"Name"`
	Config struct {
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	NetworkSettings struct {
		Networks map[string]struct {
			Gateway     string `json:"Gateway"`
			IPAddress   string `json:"IPAddress"`
			IPPrefixLen int    `json:"IPPrefixLen"`
			MacAddress  string `json:"MacAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

type VPNContainer struct {
	ID        string
	Name      string
	IPAddress string
}

func (v *VPNContainer) String() string {
	return fmt.Sprintf("Name=%s; ID=%s; IPAddress=%s", v.Name, v.ID, v.IPAddress)
}

func NewVPNContainer(networkName string, configFile string) (*VPNContainer, error) {
	// TODO: use docker library instead of exec
	id, err := exec.Command(
		"docker", "run",
		"--network", networkName,
		"--restart", "unless-stopped",
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("name=%s", networkName),
		"-d", imageName, configFile,
	).Output()
	if err != nil {
		return nil, err
	}

	// TODO: clean up if this call fails
	return NewVPNContainerFromID(string(id[:len(id)-1]))
}

func NewVPNContainerFromID(id string) (*VPNContainer, error) {
	output, err := exec.Command("docker", "inspect", id).Output()
	if err != nil {
		return nil, err
	}

	inspect := []ContainerInspectOutput{}
	if err := json.Unmarshal(output, &inspect); err != nil {
		return nil, err
	}
	networkName := inspect[0].Config.Labels["name"]

	return &VPNContainer{
		ID:        inspect[0].ID,
		Name:      inspect[0].Name,
		IPAddress: inspect[0].NetworkSettings.Networks[networkName].IPAddress,
	}, nil
}

func (v *VPNContainer) Close() error {
	return exec.Command("docker", "rm", "-f", v.ID).Run()
}
