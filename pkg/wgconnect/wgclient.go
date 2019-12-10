package wgconnect

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gosuri/uilive"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func connectToWg(ctx context.Context, opts ConnectOpts) error {
	client, err := wgctrl.New()
	if err != nil {
		return errors.Wrap(err, "failed to create wg client")
	}

	privateKey, err := wgtypes.ParseKey(opts.Interface.PrivateKey)
	if err != nil {
		return errors.Wrap(err, "failed to parse private key")
	}

	publicKey, err := wgtypes.ParseKey(opts.Peer.PublicKey)
	if err != nil {
		return errors.Wrap(err, "failed to parse public key")
	}

	allowedIPs := []net.IPNet{}
	for _, ip := range opts.Peer.AllowedIPs {
		if _, netip, err := net.ParseCIDR(ip); err == nil {
			allowedIPs = append(allowedIPs, *netip)
		} else {
			logrus.Warnf("unable to parse allowed ip cidr - it will be ignored: %s", ip)
		}
	}

	udpaddr, err := net.ResolveUDPAddr("udp", opts.Peer.Endpoint)
	if err != nil {
		return errors.Wrap(err, "failed to parse endpoint address")
	}

	client.ConfigureDevice("wg0", wgtypes.Config{
		PrivateKey: &privateKey,
		Peers: []wgtypes.PeerConfig{
			wgtypes.PeerConfig{
				PublicKey:  publicKey,
				AllowedIPs: allowedIPs,
				Endpoint:   udpaddr,
			},
		},
	})

	startTime := time.Now()
	connected := false

	go func() {
		writer := uilive.New()
		writer.Start()
		for {
			select {
			case <-ctx.Done():
				writer.Stop()
				return
			case <-time.After(1 * time.Second):
				device, err := client.Device("wg0")
				if err != nil {
					fmt.Fprintf(writer, "failed to get wireguard interface: %v\n", err)
				} else {
					peer := device.Peers[0]
					now := time.Now()
					connected = peer.LastHandshakeTime != time.Time{}
					fmt.Fprintf(writer, "connected: %v\n", connected)
					fmt.Fprintf(writer, "duration: %.0f seconds\n", now.Sub(startTime).Seconds())
					fmt.Fprintf(writer, "last handshake: %0.f seconds ago\n", now.Sub(peer.LastHandshakeTime).Seconds())
					fmt.Fprintf(writer, "sent: %v bytes\n", peer.TransmitBytes)
					fmt.Fprintf(writer, "received: %v bytes\n", peer.ReceiveBytes)
				}
			}
		}
	}()

	logrus.Debug("client wireguard device configured")

	link, err := netlink.LinkByName("wg0")
	if err != nil {
		return errors.Wrap(err, "failed to find wireguard interface")
	}

	if err := netlink.LinkSetUp(link); err != nil {
		logrus.Error(errors.Wrap(err, "failed to bring wireguard interface up"))
	}

	linkaddr, err := netlink.ParseAddr(opts.Interface.Address)
	if err != nil {
		return errors.Wrap(err, "failed to parse wireguard interface ip address")
	}

	if err := netlink.AddrAdd(link, linkaddr); err != nil {
		return errors.Wrap(err, "failed to set ip address of wireguard interface")
	}

	for {
		if connected {
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
			break
		}
		time.Sleep(1 * time.Second)
	}

	logrus.Debug("client ready")
	select {
	case <-ctx.Done():
	}

	logrus.Debug("shutting down client...")

	client.Close()

	return nil
}
