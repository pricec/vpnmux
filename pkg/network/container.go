package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

const imageName = "openvpn-client"

type ContainerInspectOutput struct {
	ID    string   `json:"Id"`
	Args  []string `json:"Args"`
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

type Container struct {
	Config       *openvpn.Config
	ID           string
	DockerID     string
	Name         string
	RouteTableID int
	IPAddress    string
}

func NewContainer(id string, cfg *openvpn.Config) (*Container, error) {
	// TODO: use docker library instead of exec
	routeTableID, err := unusedRouteTableID()
	if err != nil {
		return nil, err
	}

	err = exec.Command(
		"docker", "run",
		"--network", id,
		"--restart", "unless-stopped",
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"-v", fmt.Sprintf("%s:/etc/openvpn/config", cfg.Dir),
		"-w", "/etc/openvpn/config",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("id=%s", id),
		"--label", fmt.Sprintf("config-id=%s", cfg.ID),
		"--label", fmt.Sprintf("route-table-id=%d", routeTableID),
		"-d", imageName, "openvpn.conf",
	).Run()
	if err != nil {
		return nil, err
	}
	// TODO: clean up if this fails?
	return NewContainerFromID(id)
}

func NewContainerFromID(id string) (*Container, error) {
	output, err := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("label=id=%s", id)).Output()
	if err != nil {
		return nil, err
	}

	dockerID := string(output[:len(output)-1])
	output, err = exec.Command("docker", "inspect", dockerID).Output()
	if err != nil {
		return nil, err
	}

	inspect := []ContainerInspectOutput{}
	if err := json.Unmarshal(output, &inspect); err != nil {
		return nil, err
	}

	routeTableID, err := strconv.Atoi(inspect[0].Config.Labels["route-table-id"])
	if err != nil {
		return nil, err
	}

	cfg, err := openvpn.NewConfigFromID(inspect[0].Config.Labels["config-id"])
	if err != nil {
		return nil, err
	}

	v := &Container{
		Config:       cfg,
		ID:           inspect[0].ID,
		DockerID:     dockerID,
		Name:         inspect[0].Name,
		IPAddress:    inspect[0].NetworkSettings.Networks[id].IPAddress,
		RouteTableID: routeTableID,
	}

	if err := v.configureRouting(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Container) configureRouting() error {
	exists, ip, err := defaultRouteForTable(v.RouteTableID)
	if err != nil {
		return err
	}

	if exists && ip == v.IPAddress {
		return nil
	} else if exists {
		if err := exec.Command("ip", "route", "del", "default", "table", strconv.Itoa(v.RouteTableID)).Run(); err != nil {
			return err
		}
	}
	return exec.Command("ip", "route", "add", "default", "via", v.IPAddress, "table", strconv.Itoa(v.RouteTableID)).Run()
}

func (v *Container) Close() error {
	var result error

	if err := exec.Command("docker", "rm", "-f", v.DockerID).Run(); err != nil {
		result = multierror.Append(result, err)
	}

	// TODO: clean up routing rules?

	return result
}