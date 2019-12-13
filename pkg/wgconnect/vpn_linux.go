// +build linux

package wgconnect

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/place1/wg-embed/pkg/wgembed"
	"github.com/sirupsen/logrus"
)

func vpn(iface *wgembed.WireGuardInterface) error {
	upstreams := []string{}
	for _, upstream := range iface.Config().Interface.DNS {
		upstreams = append(upstreams, fmt.Sprintf("nameserver %s\n", upstream))
	}

	if len(upstreams) > 0 {
		cmd := exec.Command("resolvconf", "-a", iface.Name(), "-m", "0", "-x")
		cmd.Stdin = strings.NewReader(strings.Join(upstreams, ""))
		if err := cmd.Run(); err != nil {
			logrus.Fatal(errors.Wrap(err, "failed to set DNS configuration"))
		}
		logrus.Debug("set DNS config")
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
	logrus.Debug("set iptables rules")

	return nil
}
