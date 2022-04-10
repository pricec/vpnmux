package network

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

const imageName = "openvpn-client"

var (
	reDefaultRoute = regexp.MustCompile(`via [0-9.]+`)
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

func unusedRouteTableID() (int, error) {
	// return an unused route table ID in the range [1,252]
	for i := 1; i < 253; i = i + 1 {
		output, err := exec.Command("ip", "route", "show", "table", strconv.Itoa(i)).Output()
		if err != nil {
			return 0, err
		} else if len(output) == 0 {
			return i, nil
		}
	}
	return 0, fmt.Errorf("all routing table IDs seem to be in use")
}

func defaultRouteForTable(tableID int) (bool, string, error) {
	output, err := exec.Command("ip", "route", "show", "table", strconv.Itoa(tableID), "default").Output()
	if err != nil {
		return false, "", err
	} else if len(output) == 0 {
		return false, "", nil
	}

	parts := strings.Split(string(output[:len(output)-1]), "\n")
	switch len(parts) {
	case 0:
		return false, "", nil
	case 1:
	default:
		log.Panicf("found %d default routes for table %d", len(parts), tableID)
	}

	s := reDefaultRoute.FindString(parts[0])
	if s == "" {
		return false, "", fmt.Errorf("string didn't match regexp")
	}

	parts = strings.Split(s, " ")
	if len(parts) != 2 {
		return false, "", fmt.Errorf("unexpected regexp match")
	}
	return true, parts[1], nil
}
