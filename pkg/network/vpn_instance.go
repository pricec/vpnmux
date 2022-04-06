package network

import (
	"github.com/hashicorp/go-multierror"
)

type VPNInstance struct {
	Network   *VPNNetwork
	Container *VPNContainer
}

func NewVPNInstance(name string, config string) (*VPNInstance, error) {
	net, err := NewVPNNetwork(name)
	if err != nil {
		return nil, err
	}

	ctr, err := NewVPNContainer(name, config)
	if err != nil {
		// TODO: handle returned error?
		net.Close()
		return nil, err
	}

	return &VPNInstance{
		Network:   net,
		Container: ctr,
	}, nil
}

func (v *VPNInstance) Close() error {
	var result error

	// TODO: do i need to check err or can i just Append nil errors?
	if err := v.Container.Close(); err != nil {
		multierror.Append(result, err)
	}

	if err := v.Network.Close(); err != nil {
		multierror.Append(result, err)
	}

	return result
}
