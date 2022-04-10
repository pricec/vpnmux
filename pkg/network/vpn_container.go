package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

type VPNContainer struct {
	cfg          *openvpn.Config
	ID           string
	Name         string
	Config       string
	RouteTableID int
	IPAddress    string
}

func (v *VPNContainer) String() string {
	return fmt.Sprintf("Name=%s; ID=%s; IPAddress=%s", v.Name, v.ID, v.IPAddress)
}

func NewVPNContainer(networkName, host, user, pass string) (*VPNContainer, error) {
	// TODO: use docker library instead of exec
	routeTableID, err := unusedRouteTableID()
	if err != nil {
		return nil, err
	}

	// TODO: cfg will be orphaned?
	cfg, err := openvpn.NewConfig(networkName, host, user, pass)
	if err != nil {
		return nil, err
	}

	id, err := exec.Command(
		"docker", "run",
		"--network", networkName,
		"--restart", "unless-stopped",
		"--cap-add", "NET_ADMIN",
		"--device", "/dev/net/tun",
		"-v", fmt.Sprintf("%s:/etc/openvpn/config", cfg.Dir),
		"-w", "/etc/openvpn/config",
		"--label", fmt.Sprintf("%s=%s", labelKey, labelValue),
		"--label", fmt.Sprintf("name=%s", networkName),
		"--label", fmt.Sprintf("route-table-id=%d", routeTableID),
		"-d", imageName, "openvpn.conf",
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
	routeTableID, err := strconv.Atoi(inspect[0].Config.Labels["route-table-id"])
	if err != nil {
		return nil, err
	}

	cfg, err := openvpn.NewConfigFromName(networkName)
	if err != nil {
		return nil, err
	}

	v := &VPNContainer{
		cfg:          cfg,
		ID:           inspect[0].ID,
		Name:         inspect[0].Name,
		Config:       inspect[0].Args[0],
		IPAddress:    inspect[0].NetworkSettings.Networks[networkName].IPAddress,
		RouteTableID: routeTableID,
	}

	if err := v.configureRouting(); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *VPNContainer) configureRouting() error {
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

func (v *VPNContainer) Close() error {
	var result error

	if err := v.cfg.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := exec.Command("docker", "rm", "-f", v.ID).Run(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
