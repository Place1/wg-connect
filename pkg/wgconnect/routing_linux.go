// +build linux

package wgconnect

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

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
		Priority:  1,
	})
	if err != nil {
		return errors.Wrap(err, "failed to add network route for wireguard traffic")
	}

	return nil
}

func setDNS(name string, upstreams ...string) error {
	cmd := exec.Command("resolvconf", "-a", "wg0", "-m", "0", "-x")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get cmd stdin pipe")
	}

	lines := []string{}
	for _, upstream := range upstreams {
		lines = append(lines, fmt.Sprintf("namespace %s\n", upstream))
	}

	if _, err := stdin.Write([]byte(strings.Join(lines, ""))); err != nil {
		return errors.Wrap(err, "failed to write to cmd stdin")
	}
	stdin.Close()

	return cmd.Run()
}
