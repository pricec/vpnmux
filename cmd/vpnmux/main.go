package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pricec/vpnmux/pkg/network"
)

func main() {
	n, err := network.NewVPNNetwork("thing")
	if err != nil {
		fmt.Printf("error creating network: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("VPN network: %v\n", n.String())

	c, err := network.NewVPNContainer(n.Name, "us-sea.prod.surfshark.com_udp.ovpn")
	if err != nil {
		fmt.Printf("error creating container: %v\n", err)
		os.Exit(-1)
	}

	rt, err := network.NewVPNRoutingTable(100, c.IPAddress)
	if err != nil {
		fmt.Printf("error setting up routing table: %v\n", err)
		os.Exit(-1)
	}

	if err := rt.AddSource("192.168.1.60"); err != nil {
		fmt.Printf("error adding source to routing table: %v\n, err")
	}

	<-time.After(10 * time.Second)

	if err := rt.Close(); err != nil {
		fmt.Printf("error closing routing table: %v\n", err)
	}

	if err := c.Close(); err != nil {
		fmt.Printf("error closing container: %v\n", err)
	}

	if err := n.Close(); err != nil {
		fmt.Printf("error closing network: %v\n", err)
	}
}
