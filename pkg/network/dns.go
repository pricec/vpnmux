package network

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"

	multierror "github.com/hashicorp/go-multierror"
)

type DNSRouter struct {
	Mark         string
	LocalSubnet  string
	Gateway      string
	RouteTableID int
}

// TODO: how to prevent multiple creation? It only seems possible to use
//       iptables -C if the mark value is known, but what if someone else
//       instantiated us with a different mark?
func NewDNSRouter(ctx context.Context, mark, localSubnet string) (*DNSRouter, error) {
	r := &DNSRouter{
		Mark:        mark,
		LocalSubnet: localSubnet,
	}

	if err := r.ensureMark(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *DNSRouter) Close() error {
	var result error

	if err := r.iptablesCommand("D", "udp").Run(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := r.iptablesCommand("D", "tcp").Run(); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func (r *DNSRouter) ensureMark() error {
	for _, proto := range []string{"tcp", "udp"} {
		cmd := r.iptablesCommand("C", proto)
		err := cmd.Run()
		switch cmd.ProcessState.ExitCode() {
		case 0:
		case 1:
			if err := r.iptablesCommand("A", proto).Run(); err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

func (r *DNSRouter) iptablesCommand(operation, proto string) *exec.Cmd {
	return exec.Command(
		"iptables",
		"-t", "mangle",
		fmt.Sprintf("-%s", operation), "OUTPUT",
		"-p", proto,
		"--dport", "53",
		"!", "-d", r.LocalSubnet,
		"-j", "MARK",
		"--set-mark", r.Mark,
	)
}

func (r *DNSRouter) Route(via string) error {
	ids, err := routeTableIDsForSelector("fwmark", r.Mark)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return r.setupTable(via)
	}

	exists, gateway, err := defaultRouteForTable(ids[0])
	if !exists {
		// TODO: is there a better way to handle this case?
		return fmt.Errorf("no default route for table %d: %v", ids[0], err)
	}

	r.RouteTableID = ids[0]
	r.Gateway = gateway
	return nil
}

func (r *DNSRouter) setupTable(via string) error {
	// TODO: check if we are already routing via `via`
	// Create routing table, default via `via`
	rtid, err := unusedRouteTableID()
	if err != nil {
		return err
	}

	err = exec.Command(
		"ip", "route", "add",
		"default", "via", via, "table", strconv.Itoa(rtid),
	).Run()
	if err != nil {
		return err
	}

	// Create RPDB rule, lookup on r.Mark
	err = exec.Command("ip", "rule", "add", "fwmark", r.Mark, "lookup", strconv.Itoa(rtid)).Run()
	if err != nil {
		// TODO: remove default route rule
		return err
	}

	r.RouteTableID = rtid
	r.Gateway = via
	return nil
}

func (r *DNSRouter) Clear() error {
	err := exec.Command("ip", "rule", "del", "fwmark", r.Mark).Run()
	if err != nil {
		return err
	}

	err = exec.Command("ip", "route", "del", "default", "table", strconv.Itoa(r.RouteTableID)).Run()
	if err != nil {
		return err
	}

	return nil
}
