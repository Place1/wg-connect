package wgconnect

import (
	"github.com/pkg/errors"
	"github.com/place1/wg-embed/pkg/wgembed"
)

type WireGuardVPN struct {
	iface *wgembed.WireGuardInterface
}

// Connect will start a userspace wireguard interface
// and then connect to a peer using the given connect opts
func Connect(configFile string) (*WireGuardVPN, error) {
	vpn := &WireGuardVPN{}

	wg0, err := wgembed.New("wg0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wireguard interface")
	}
	vpn.iface = wg0

	if err := wg0.LoadConfig(configFile); err != nil {
		return nil, errors.Wrap(err, "failed to configure wireguard interface")
	}

	if err := VPN(wg0); err != nil {
		return nil, errors.Wrap(err, "failed to setup vpn networking")
	}

	return vpn, nil
}

func (vpn *WireGuardVPN) Iface() *wgembed.WireGuardInterface {
	return vpn.iface
}

// Close ends the VPN connection
func (vpn *WireGuardVPN) Close() error {
	return vpn.iface.Close()
}
