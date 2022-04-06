package network

import (
	"os/exec"
	"strconv"
)

const lanInterface = "lan0"
const wanInterface = "wan0"

type VPNRoutingTable struct {
	tableID   int
	sourceIPs map[string]struct{}
}

func NewVPNRoutingTable(tableID int, gatewayIP string) (*VPNRoutingTable, error) {
	// TODO: use library code
	// TODO: what if tableID already exists?
	err := exec.Command(
		"ip", "route", "add",
		"default", "via", gatewayIP,
		"table", strconv.Itoa(tableID),
	).Run()
	if err != nil {
		return nil, err
	}

	return &VPNRoutingTable{
		tableID:   tableID,
		sourceIPs: make(map[string]struct{}),
	}, nil
}

func (v *VPNRoutingTable) Close() error {
	sources := make([]string, 0)
	for source, _ := range v.sourceIPs {
		sources = append(sources, source)
	}

	for _, source := range sources {
		if err := v.RemoveSource(source); err != nil {
			// TODO: inconsistent state?
			return err
		}
	}
	return exec.Command("ip", "route", "del", "default", "table", strconv.Itoa(v.tableID)).Run()
}

func (v *VPNRoutingTable) AddSource(ip string) error {
	// TODO: add locking
	if _, ok := v.sourceIPs[ip]; ok {
		return nil
	}

	err := exec.Command("ip", "rule", "add", "from", ip, "lookup", strconv.Itoa(v.tableID)).Run()
	if err != nil {
		return err
	}

	// TODO: what if this call fails? inconsistent state
	// TODO: library code
	err = exec.Command(
		"iptables",
		"-t", "filter",
		"-A", "FORWARD",
		"-i", lanInterface,
		"-o", wanInterface,
		"-s", ip,
		"-j", "DROP",
	).Run()
	if err != nil {
		return err
	}

	v.sourceIPs[ip] = struct{}{}
	return nil
}

func (v *VPNRoutingTable) RemoveSource(ip string) error {
	if _, ok := v.sourceIPs[ip]; !ok {
		return nil
	}

	err := exec.Command("ip", "rule", "del", "from", ip, "lookup", strconv.Itoa(v.tableID)).Run()
	if err != nil {
		return err
	}

	// TODO: what if this call fails? inconsistent state
	err = exec.Command(
		"iptables",
		"-t", "filter",
		"-D", "FORWARD",
		"-i", lanInterface,
		"-o", wanInterface,
		"-s", ip,
		"-j", "DROP",
	).Run()
	if err != nil {
		return err
	}

	delete(v.sourceIPs, ip)
	return nil
}
