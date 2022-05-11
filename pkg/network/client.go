package network

import (
	"fmt"
	"os/exec"
	"strconv"
)

type Client struct {
	Address      string
	LANInterface string
	WANInterface string
}

func NewClient(address, lanInterface, wanInterface string) (*Client, error) {
	c := &Client{
		Address:      address,
		LANInterface: lanInterface,
		WANInterface: wanInterface,
	}

	if err := c.preventForwarding(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() error {
	return c.iptablesCommand("D").Run()
}

func (c *Client) iptablesCommand(operation string) *exec.Cmd {
	return exec.Command(
		"iptables",
		"-t", "filter",
		fmt.Sprintf("-%s", operation), "FORWARD",
		"-i", c.LANInterface,
		"-o", c.WANInterface,
		"-s", c.Address,
		"-j", "DROP",
	)
}

func (c *Client) preventForwarding() error {
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

func (c *Client) SetRouteTable(id int) error {
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

func (c *Client) ClearRoutes() error {
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
