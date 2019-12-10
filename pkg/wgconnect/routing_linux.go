// +build linux

package wgconnect

import (
	"net"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func ifaceUp(name string, ip string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Wrap(err, "failed to find wireguard interface")
	}

	if err := netlink.LinkSetUp(link); err != nil {
		logrus.Error(errors.Wrap(err, "failed to bring wireguard interface up"))
	}

	linkaddr, err := netlink.ParseAddr(ip)
	if err != nil {
		return errors.Wrap(err, "failed to parse wireguard interface ip address")
	}

	if err := netlink.AddrAdd(link, linkaddr); err != nil {
		return errors.Wrap(err, "failed to set ip address of wireguard interface")
	}

	return nil
}

func ifaceDefaultRoute(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Wrap(err, "failed to find wireguard interface")
	}

	_, allnetip, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return errors.Wrap(err, "failed to parse 0.0.0.0/0 cidr")
	}

	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       allnetip,
		Scope:     netlink.SCOPE_LINK,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add network route for wireguard traffic")
	}

	return nil
}
