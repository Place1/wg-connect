package wgconnect

import "github.com/place1/wg-embed/pkg/wgembed"

// VPN configures the operating system's network stack
// to use the given interface (e.g. wg0) as a VPN
func VPN(iface *wgembed.WireGuardInterface) error {
	return vpn(iface)
}
