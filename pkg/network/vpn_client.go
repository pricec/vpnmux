package network

import (
	"fmt"
	"os/exec"
)

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
