package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	reRouteTableID = regexp.MustCompile(`lookup \d+`)
)

// true iff source is currently routed to a numbered route table
// TODO: ugly implemenation can probably be improved upon
// TODO: use library code
func routeTableIDsForSource(source string) ([]int, error) {
	output, err := exec.Command("ip", "rule", "show", "from", source).Output()
	if err != nil {
		return nil, err
	}

	var result []int
	parts := strings.Split(string(output), "\n")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		s := reRouteTableID.FindString(part)
		if s == "" {
			return nil, fmt.Errorf("string didn't match regexp")
		}

		subparts := strings.Split(s, " ")
		id, err := strconv.Atoi(subparts[1])
		if err != nil {
			return nil, err
		}

		result = append(result, id)
	}
	return result, nil
}

type VPNClient struct {
	Address string
}

func NewVPNClient(address string) (*VPNClient, error) {
	c := &VPNClient{
		Address: address,
	}

	if err := c.preventForwarding(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *VPNClient) Close() error {
	return c.iptablesCommand("D").Run()
}

func (c *VPNClient) iptablesCommand(operation string) *exec.Cmd {
	return exec.Command(
		"iptables",
		"-t", "filter",
		fmt.Sprintf("-%s", operation), "FORWARD",
		"-i", lanInterface,
		"-o", wanInterface,
		"-s", c.Address,
		"-j", "DROP",
	)
}

func (c *VPNClient) preventForwarding() error {
	// Ensure packets are not forwarded from LAN -> WAN, since
	// they should be routed via one of the managed VPNs.
	// TODO: use library code instead of exec
	cmd := c.iptablesCommand("C")
	// Note that this command can return nonzero if the rule exists
	err := cmd.Run()
	switch cmd.ProcessState.ExitCode() {
	case 0:
		return nil
	case 1:
	default:
		return err
	}

	return c.iptablesCommand("A").Run()
}

func (c *VPNClient) SetRouteTable(id int) error {
	routeTableIDs, err := routeTableIDsForSource(c.Address)
	if err != nil {
		return err
	}

	found := false
	for _, routeTableID := range routeTableIDs {
		if routeTableID != id {
			err = exec.Command("ip", "rule", "del", "from", c.Address, "lookup", strconv.Itoa(routeTableID)).Run()
			if err != nil {
				return err
			}
		} else {
			found = true
		}
	}

	if !found {
		return exec.Command("ip", "rule", "add", "from", c.Address, "lookup", strconv.Itoa(id)).Run()
	}

	return nil
}

func (c *VPNClient) ClearRoutes() error {
	routeTableIDs, err := routeTableIDsForSource(c.Address)
	if err != nil {
		return err
	}

	for _, routeTableID := range routeTableIDs {
		err = exec.Command("ip", "rule", "del", "from", c.Address, "lookup", strconv.Itoa(routeTableID)).Run()
		if err != nil {
			return err
		}
	}
	return nil
}
