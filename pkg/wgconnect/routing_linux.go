// +build linux

package wgconnect

import (
	"os/exec"

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

	if err := netlink.LinkSetMTU(link, 1420); err != nil {
		return errors.Wrap(err, "failed to set wireguard mtu")
	}

	return nil
}

func ifaceDefaultRoute(name string) error {

	err := exec.Command("bash", "-c", "echo nameserver 8.8.8.8 | resolvconf -a wg0 -m 0 -x").Run()
	if err != nil {
		logrus.Fatal(err)
	}

	commands := [][]string{
		[]string{"-6", "route", "add", "::/0", "dev", "wg0", "table", "51820"},
		[]string{"-6", "rule", "add", "not", "fwmark", "51820", "table", "51820"},
		[]string{"-6", "rule", "add", "table", "main", "suppress_prefixlength", "0"},
		[]string{"-4", "route", "add", "0.0.0.0/0", "dev", "wg0", "table", "51820"},
		[]string{"-4", "rule", "add", "not", "fwmark", "51820", "table", "51820"},
		[]string{"-4", "rule", "add", "table", "main", "suppress_prefixlength", "0"},
	}

	for _, cmd := range commands {
		if err := exec.Command("ip", cmd...).Run(); err != nil {
			logrus.Fatal(err)
		}
	}

	return nil
}
