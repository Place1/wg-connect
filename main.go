package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/place1/wg-embed/pkg/wgembed"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uilive"
	"github.com/place1/wg-connect/pkg/wgconnect"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app    = kingpin.New("wg-connect", "a cli app to connect to a wireguard VPN server using a userspace wireguard implementation")
	config = app.Arg("config", "a wireguard configuration file").Required().String()
	debug  = app.Flag("debug", "enable verbose logging").Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	vpn, err := wgconnect.Connect(*config)
	if err != nil {
		logrus.Error(err)
	}

	go monitor(vpn.Iface())

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	}

	logrus.Debug("shutting down...")

	vpn.Close()
}

func monitor(iface *wgembed.WireGuardInterface) error {
	writer := uilive.New()
	writer.Start()
	for {
		time.Sleep(1 * time.Second)
		device, err := iface.Device()
		if err != nil {
			fmt.Fprintf(writer, "failed to get wireguard interface: %v\n", err)
		} else {
			if len(device.Peers) > 0 {
				peer := device.Peers[0]
				connected := peer.LastHandshakeTime != time.Time{}
				fmt.Fprintf(writer, "connected:      %v\n", connected)
				fmt.Fprintf(writer, "last handshake: %s\n", humanize.Time(peer.LastHandshakeTime))
				fmt.Fprintf(writer, "sent:           %v\n", humanize.Bytes(uint64(peer.TransmitBytes)))
				fmt.Fprintf(writer, "received:       %v\n", humanize.Bytes(uint64(peer.ReceiveBytes)))
			}
		}
	}
}
