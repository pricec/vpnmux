package main

import (
	"fmt"
	"os"

	"github.com/pricec/vpnmux/pkg/network"
)

func main() {
	n, err := network.NewVPNNetwork("thing")
	if err != nil {
		fmt.Printf("error creating network: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("VPN network: %v\n", n.String())

	c, err := network.NewVPNContainer(n.ID, "us-sea.prod.surfshark.com_udp.ovpn")
	if err != nil {
		fmt.Printf("error creating container: %v\n", err)
	}

	if err := c.Close(); err != nil {
		fmt.Printf("error closing container: %v\n", err)
	}

	if err := n.Close(); err != nil {
		fmt.Printf("error closing network: %v\n", err)
	}
}
