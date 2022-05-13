package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

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

func NewContainer(id, image, subnet string, cfg *openvpn.Config) (*Container, error) {
	// TODO: use docker library instead of exec
	routeTableID, err := unusedRouteTableID()
	if err != nil {
		return nil, err
	}

	output, err := exec.Command(
		"docker", "run",
		"--network", id,
		"--restart", "unless-stopped",
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"-v", fmt.Sprintf("%s:/etc/openvpn/config", cfg.Dir),
		"-w", "/etc/openvpn/config",
		"-e", fmt.Sprintf("LOCAL_SUBNET_CIDR=%s", subnet),
		"--pull", "always",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("id=%s", id),
		"--label", fmt.Sprintf("config-id=%s", cfg.ID),
		"--label", fmt.Sprintf("route-table-id=%d", routeTableID),
		"-d", image, "openvpn.conf",
	).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("running container: %w; %v", err, string(output))
	}
	// TODO: clean up if this fails?
	return NewContainerFromID(id)
}

func NewContainerFromID(id string) (*Container, error) {
	// TODO: what if output is empty? we crash.
	output, err := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("label=id=%s", id)).Output()
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	dockerID := string(output[:len(output)-1])
	output, err = exec.Command("docker", "inspect", dockerID).Output()
	if err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}

	inspect := []ContainerInspectOutput{}
	if err := json.Unmarshal(output, &inspect); err != nil {
		return nil, fmt.Errorf("unmarshaling container inspect: %w", err)
	}

	routeTableID, err := strconv.Atoi(inspect[0].Config.Labels["route-table-id"])
	if err != nil {
		return nil, fmt.Errorf("parsing route-table-id: %w", err)
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
			return fmt.Errorf("ip route del default: %w", err)
		}
	}

	output, err := exec.Command("ip", "route", "add", "default", "via", v.IPAddress, "table", strconv.Itoa(v.RouteTableID)).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return fmt.Errorf("ip route add default: %w", err)
	}
	return nil
}

func (v *Container) Close() error {
	var result error

	if err := exec.Command("docker", "rm", "-f", v.DockerID).Run(); err != nil {
		result = multierror.Append(result, err)
	}

	// TODO: clean up routing rules?

	return result
}
