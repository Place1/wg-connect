package wgconnect

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uilive"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	logrus.Debug("client wireguard device configured")

	if err := ifaceUp("wg0", opts.Interface.Address); err != nil {
		return errors.Wrap(err, "unable to bring wireguard interface up")
	}

	logrus.Debug("wireguard interface up")

	updateInterval := 500 * time.Millisecond
	connected := false

	go func() {
		writer := uilive.New()
		writer.Start()
		for {
			select {
			case <-ctx.Done():
				writer.Stop()
				return
			case <-time.After(updateInterval):
				device, err := client.Device("wg0")
				if err != nil {
					fmt.Fprintf(writer, "failed to get wireguard interface: %v\n", err)
				} else {
					peer := device.Peers[0]
					connected = peer.LastHandshakeTime != time.Time{}
					fmt.Fprintf(writer, "connected:      %v\n", connected)
					fmt.Fprintf(writer, "last handshake: %s\n", humanize.Time(peer.LastHandshakeTime))
					fmt.Fprintf(writer, "sent:           %v\n", humanize.Bytes(uint64(peer.TransmitBytes)))
					fmt.Fprintf(writer, "received:       %v\n", humanize.Bytes(uint64(peer.ReceiveBytes)))
				}
			}
		}
	}()

	for {
		if connected {
			if err := ifaceDefaultRoute("wg0"); err != nil {
				return errors.Wrap(err, "failed to set default route wireguard")
			}
			break
		}
		time.Sleep(updateInterval)
	}

	logrus.Debug("client ready")
	select {
	case <-ctx.Done():
	}

	logrus.Debug("shutting down client...")

	client.Close()

	return nil
}
